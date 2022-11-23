package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"princepereira/TcpClientServer/util"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var allFailedCons map[int][]util.ConnInfo // Stores entire failed info
var failedCons *failedConnsStruct         // Stores failed info for each iteration
var serverInfoMap *sync.Map

type failedConnsStruct struct {
	failedCons []util.ConnInfo
	mutex      *sync.Mutex
}

func (failedConns *failedConnsStruct) append(connInfo util.ConnInfo) {
	failedConns.mutex.Lock()
	defer failedConns.mutex.Unlock()
	failedConns.failedCons = append(failedConns.failedCons, connInfo)
}

func (failedConns *failedConnsStruct) string() string {
	failedConns.mutex.Lock()
	defer failedConns.mutex.Unlock()
	str := ""
	for _, v := range failedConns.failedCons {
		failure, err := json.MarshalIndent(v, "", "  ")
		if err == nil {
			str = str + "\n" + string(failure)
		}
	}
	return str
}

func (failedConns *failedConnsStruct) size() int {
	failedConns.mutex.Lock()
	defer failedConns.mutex.Unlock()
	return len(failedConns.failedCons)
}

func main() {

	args, err := util.ValidateArgs()
	if err != nil {
		util.ClientHelp()
		log.Println(err)
		return
	}

	if args[util.AtribHelp] == util.ConstTrue {
		util.ClientHelp()
		return
	}

	if args[util.AtribVersion] == util.ConstTrue {
		fmt.Println(util.Version)
		return
	}

	if err := util.ValidateValues("client", args); err != nil {
		util.ClientHelp()
		log.Println(err)
		return
	}

	util.PrintClientBanner(args)

	conns, _ := strconv.Atoi(args[util.AtribCons])
	proto := args[util.AtribProto]
	iter, _ := strconv.Atoi(args[util.AtribIterations])
	util.SetMaxDropThreshold(args[util.AtribMaxDropThreshold])

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(chanSignal, cancel)

	var address string

	if strings.Contains(args[util.AtribIpAddr], ":") {
		// For IPv6 support
		address = fmt.Sprintf("[%s]:%s", args[util.AtribIpAddr], args[util.AtribPort])
	} else {
		address = fmt.Sprintf("%s:%s", args[util.AtribIpAddr], args[util.AtribPort])
	}

	allFailedCons = make(map[int][]util.ConnInfo)

	for turn := 1; turn <= iter && util.NoExitClient; turn++ {

		serverInfoMap = new(sync.Map)
		failedCons = &failedConnsStruct{failedCons: make([]util.ConnInfo, 0), mutex: &sync.Mutex{}}

		log.Printf("\n\n######=========  ITERATION : %d STARTED =========######\n\n", turn)

		switch proto {
		case util.ConstTCP:
			invokeTcpClient(proto, address, conns, turn, args, wg, ctx)
		case util.ConstUDP:
			invokeUdpClient(proto, address, conns, turn, args, wg, ctx)
		default:
			log.Fatal("No Proto defined, hence exiting...")
		}

		time.Sleep(3 * time.Second)

		failedConnCount := failedCons.size()
		passedConnCount := conns - failedConnCount
		fmt.Printf("\n\n\n#======= ConnectionsSucceded:%d, ConnectionsFailed:%d , Iteration:%d \n", passedConnCount, failedConnCount, turn)

		if failedConnCount != 0 {
			str := fmt.Sprintf("\n#======= Iteration : %d, No: of failed connections : %d", turn, failedConnCount)
			str = str + "\n\nFailed connections : \n\n"
			str = str + "=============================\n"
			str = str + "\n" + failedCons.string()
			str = str + "\n\n=============================\n"
			// str = str + fmt.Sprintf("\nClient Details : %s", util.GetIPAddress())
			log.Println(str)
			allFailedCons[turn] = failedCons.failedCons
		} else {
			log.Printf("\n\n######========= ALL CONNECTIONS ARE CLOSED GRACEFULLY FOR ITERATION : %d =========######\n\n", turn)
		}

	}

	log.Println("\n\nExiting TCP Client ...")
}

func invokeUdpClient(proto, address string, conns, iter int, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {

		time.Sleep(2 * time.Second)
		if ctx.Err() != nil {
			log.Println(util.ConnTerminatedSuccessMsg, "UDP Connection create terminated by ctrl+c signal. ", util.ConnTerminatedMsg)
			return
		}

		clientName := util.GetClientName(proto, iter, i)

		udpServer, err := net.ResolveUDPAddr(proto, address)
		if err != nil {
			storeConnFailure(clientName, address, "", err.Error(), i, 0, nil, nil, false)
			continue
		}

		c, err := net.DialUDP(proto, nil, udpServer)
		if err != nil {
			storeConnFailure(clientName, address, "", err.Error(), i, 0, nil, nil, false)
			continue
		}

		log.Println("#===== UDPClient Port Opened : ", clientName, ", LocalAddress : ", c.LocalAddr().String(), ",  RemoteAddress : ", c.RemoteAddr().String(), " ======#")
		connMap[clientName] = c
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons.failedCons))

	for clientName, con := range connMap {
		log.Println("#===== Starting ", clientName, " ======#")
		go startUdpClient(clientName, address, con, args, wg, ctx)
	}

	wg.Wait()
}

func invokeTcpClient(proto, address string, conns, iter int, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	disableKeepAlive, _ := strconv.ParseBool(args[util.AtribDisableKeepAlive])
	keepAliveTimeOut, _ := strconv.Atoi(args[util.AtribTimeoutKeepAlive])

	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {

		time.Sleep(3 * time.Second)
		if ctx.Err() != nil {
			log.Println(util.ConnTerminatedSuccessMsg, "TCP Connection create terminated by ctrl+c signal. ", util.ConnTerminatedMsg)
			return
		}

		clientName := util.GetClientName(proto, iter, i)

		c, err := net.DialTimeout(proto, address, util.DialTimeout)
		if err != nil {
			storeConnFailure(clientName, address, "", err.Error(), i, 0, nil, nil, false)
			continue
		}

		if tc, ok := c.(*net.TCPConn); ok {
			tc.SetKeepAlive(!disableKeepAlive)
			if !disableKeepAlive {
				keepAliveTimeOutInt := time.Duration(keepAliveTimeOut) * time.Millisecond // Negative time will disable keepAlive
				tc.SetKeepAlivePeriod(keepAliveTimeOutInt)
			}
		}

		log.Println("#===== TCPClient Connected : ", clientName, ", LocalAddress : ", c.LocalAddr().String(), ",  RemoteAddress : ", c.RemoteAddr().String(), " ======#")
		connMap[clientName] = c
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons.failedCons))

	for clientName, con := range connMap {
		log.Println("#===== Starting ", clientName, " ======#")
		go startTcpClient(clientName, address, con, args, wg, ctx)
	}

	wg.Wait()
}

func storeConnFailure(clientName, remoteAddress, request, reason string, i, packetsDropped int, con net.Conn, cancelPacketTracker context.CancelFunc, exit bool) {
	t := time.Now()
	failedTime := t.Format(time.RFC3339)
	var serverInfo string
	if info, ok := serverInfoMap.Load(clientName); ok {
		serverInfo = info.(string)
	}

	if cancelPacketTracker != nil {
		cancelPacketTracker()
	}

	var localAddress string
	if con != nil {
		localAddress = con.LocalAddr().String()
		con.Close()
	}

	connInfo := util.ConnInfo{
		ClientName:    clientName,
		LocalAddess:   localAddress,
		RemoteAddress: remoteAddress,
		ServerInfo:    serverInfo,
		RequestInfo:   request,
		FailedReason:  reason,
		FailedTime:    failedTime,
	}

	failedCons.append(connInfo)

	if exit {
		log.Printf("%s, Exiting... Packets Dropped : %d . %s \n\nConnectionInfo : %v \n", util.ConnTerminatedFailedMsg, packetsDropped, util.ConnTerminatedMsg, connInfo)
	} else {
		log.Printf("\n#===== Failed to connect to : %s by client : %s . %s ======#\n\nConnectionInfo : %v \n\n", remoteAddress, clientName, util.ConnTerminatedMsg, connInfo)
	}
}

func serverMsghandler(conn net.Conn, clientName string, counter *int32, delay int, contextPacketTracker context.Context) {
	firstMsg := true
	var receivedMsg string
	waitChan := make(chan bool)

	s := bufio.NewScanner(conn)
	go packetTracker(clientName, delay, waitChan, conn, contextPacketTracker)

	for s.Scan() {

		receivedMsg = string(s.Text())
		log.Println("<<<<==== Received Message : " + receivedMsg)
		waitChan <- true

		if firstMsg {
			// Just storing the information
			firstMsg = false
			serverInfoMap.Store(clientName, receivedMsg)
		}

		if strings.Contains(receivedMsg, util.QuitMsg) {
			atomic.StoreInt32(counter, 1)
			conn.Close()
			log.Println(util.ConnTerminatedSuccessMsg, "Received quit connection from server : ", conn.RemoteAddr().String(), " , hence closing the connection... ", util.ConnTerminatedMsg)
			return
		}

	}
}

func startTcpClient(clientName, remoteAddr string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()

	var counter int32 = 0
	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])
	ctxPacketTracker, cancelPacketTracker := context.WithCancel(context.Background())

	go serverMsghandler(c, clientName, &counter, delay, ctxPacketTracker)

	dropCounter, resetCounter := 0, 0

	for i := 1; i <= requests; i++ {
		if atomic.LoadInt32(&counter) == 1 {
			// Connection is closed by server sent "Quit Message"
			cancelPacketTracker()
			return
		}
		msgSent := clientName + "- Req-" + strconv.Itoa(i) + "\n"
		_, sendErr := c.Write([]byte(msgSent))
		if sendErr != nil {
			if strings.Contains(sendErr.Error(), util.ErrMsgListenClosed) {
				// Error message is already handled in server handler
				cancelPacketTracker()
				return
			}
			if util.IsConnClosed(sendErr.Error()) {
				storeConnFailure(clientName, remoteAddr, msgSent, sendErr.Error(), i, dropCounter, c, cancelPacketTracker, true)
				return
			}
			log.Println("#====== TCP Send Error : ", sendErr.Error())
			dropCounter++
			resetCounter++
			log.Println("Packet dropped with ", clientName, " request : ", i)
			if resetCounter == util.MaxDropPackets {
				storeConnFailure(clientName, remoteAddr, msgSent, "connection is broken.", i, dropCounter, c, cancelPacketTracker, true)
				return
			}
		} else {
			resetCounter = 0
		}

		log.Println("====>>>> TCP Request Sent : " + msgSent)

		if ctx.Err() != nil {
			log.Println(util.ConnTerminatedSuccessMsg, "Connection terminated by ctrl+c signal : ", c.RemoteAddr().String(), util.ConnTerminatedMsg)
			cancelPacketTracker()
			return
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	cancelPacketTracker()
	c.Close()
}

func startUdpClient(clientName, remoteAddr string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()

	var counter int32 = 0
	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])
	ctxPacketTracker, cancelPacketTracker := context.WithCancel(context.Background())

	go serverMsghandler(c, clientName, &counter, delay, ctxPacketTracker)

	dropCounter, resetCounter := 0, 0

	for i := 1; i <= requests; i++ {

		msgSent := clientName + "- Req-" + strconv.Itoa(i) + "\n"

		_, sendErr := c.Write([]byte(msgSent))
		if sendErr != nil {
			log.Println("#====== UDP Send Error : ", sendErr.Error())
			if strings.Contains(sendErr.Error(), util.ErrMsgListenClosed) {
				// Error message is already handled in server handler
				return
			}
			dropCounter++
			resetCounter++
			log.Println("Packet dropped with ", clientName, " request : ", i)
			if resetCounter == util.MaxDropPackets {
				storeConnFailure(clientName, remoteAddr, msgSent, "connection is broken.", i, dropCounter, c, cancelPacketTracker, true)
				return
			}
		} else {
			resetCounter = 0
		}

		log.Println("====>>>> UDP Request Sent : " + msgSent)

		if ctx.Err() != nil {
			log.Println(util.ConnTerminatedSuccessMsg, "Connection terminated by ctrl+c signal : ", c.RemoteAddr().String(), util.ConnTerminatedMsg)
			return
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)

	}

	c.Close()

}

func handleCtrlC(c chan os.Signal, cancel context.CancelFunc) {
	<-c
	// handle ctrl+c event here
	log.Println("#===== Ctrl + C called ... Exiting in few seconds")
	util.NoExitClient = false
	cancel()
}

func packetTracker(clientName string, delay int, waitChan chan bool, conn net.Conn, contxt context.Context) {

	extraWaitTime := 5 * time.Second
	totalWaitTime := time.Duration(delay)*time.Millisecond + extraWaitTime
	dropCounter, resetCounter := 0, 0

	// defer close(waitChan)

	for {
		select {
		case <-waitChan:
			resetCounter++
			if resetCounter >= util.MaxDropThreshold {
				resetCounter = 0
				dropCounter = 0
			}
		case <-time.After(totalWaitTime):
			dropCounter++
			log.Println("Packet timeout : ", conn.RemoteAddr().String())
			if dropCounter >= util.MaxDropThreshold {
				storeConnFailure(clientName, conn.RemoteAddr().String(), "", "Connection timed out.", 0, dropCounter, conn, nil, true)
				conn.Close()
				return
			}
		case <-contxt.Done():
			return
		}
	}

}
