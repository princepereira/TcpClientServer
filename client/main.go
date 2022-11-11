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

var failedCons []util.ConnInfo
var serverInfoMap *sync.Map

func main() {

	args, err := util.ValidateArgs()
	if err != nil {
		util.ClientHelp()
		log.Println(err)
		return
	}

	if args[util.AtribHelp] == "true" {
		util.ClientHelp()
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

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(chanSignal, cancel)

	address := fmt.Sprintf("%s:%s", args[util.AtribIpAddr], args[util.AtribPort])

	for turn := 1; turn <= iter && util.NoExitClient; turn++ {

		serverInfoMap = new(sync.Map)
		failedCons = make([]util.ConnInfo, 0)

		log.Printf("\n\n######=========  ITERATION : %d STARTED =========######\n\n", turn)

		switch proto {
		case util.ConstTCP:
			invokeTcpClient(proto, address, conns, turn, args, wg, ctx)
		case util.ConstUDP:
			invokeUdpClient(proto, address, conns, turn, args, wg, ctx)
		default:
			log.Fatal("No Proto defined, hence exiting...")
		}

		if len(failedCons) != 0 {
			str := fmt.Sprintf("\n\n\n#======= Iteration : %d, No: of failed connections : %d", turn, len(failedCons))
			str = str + "\n\nFailed connections : \n\n"
			str = str + "=============================\n"
			for _, v := range failedCons {
				failure, err := json.MarshalIndent(v, "", "  ")
				if err == nil {
					str = str + "\n" + string(failure)
				}
			}
			str = str + "\n\n=============================\n"
			// str = str + fmt.Sprintf("\nClient Details : %s", util.GetIPAddress())
			log.Println(str)
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
			storeConnFailure(clientName, address, "", err.Error(), i, 0, nil, false)
			continue
		}

		c, err := net.DialUDP(proto, nil, udpServer)
		if err != nil {
			storeConnFailure(clientName, address, "", err.Error(), i, 0, nil, false)
			continue
		}

		log.Println("#===== UDPClient Connected : ", clientName, ", LocalAddress : ", c.LocalAddr().String(), ",  RemoteAddress : ", c.RemoteAddr().String(), " ======#")
		connMap[clientName] = c
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons))

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

		time.Sleep(2 * time.Second)
		if ctx.Err() != nil {
			log.Println(util.ConnTerminatedSuccessMsg, "TCP Connection create terminated by ctrl+c signal. ", util.ConnTerminatedMsg)
			return
		}

		clientName := util.GetClientName(proto, iter, i)

		c, err := net.DialTimeout(proto, address, util.DialTimeout)
		if err != nil {
			storeConnFailure(clientName, address, "", err.Error(), i, 0, nil, false)
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

	wg.Add(conns - len(failedCons))

	for clientName, con := range connMap {
		log.Println("#===== Starting ", clientName, " ======#")
		go startTcpClient(clientName, address, con, args, wg, ctx)
	}

	wg.Wait()
}

func storeConnFailure(clientName, remoteAddress, request, reason string, i, packetsDropped int, con net.Conn, exit bool) {
	t := time.Now()
	failedTime := t.Format(time.RFC3339)
	var serverInfo string
	if info, ok := serverInfoMap.Load(clientName); ok {
		serverInfo = info.(string)
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

	failedCons = append(failedCons, connInfo)
	if exit {
		log.Printf("%s, Exiting... Packets Dropped : %d . %s \n\nConnectionInfo : %v \n", util.ConnTerminatedFailedMsg, packetsDropped, util.ConnTerminatedMsg, connInfo)
	} else {
		log.Printf("\n#===== Failed to connect to : %s by client : %s . %s ======#\n\nConnectionInfo : %v \n\n", remoteAddress, clientName, util.ConnTerminatedMsg, connInfo)
	}
}

func serverMsghandler(conn net.Conn, clientName string, counter *int32) {
	firstMsg := true
	s := bufio.NewScanner(conn)
	for s.Scan() {
		receivedMsg := string(s.Text())
		if firstMsg {
			// Just storing the information
			firstMsg = false
			serverInfoMap.Store(clientName, receivedMsg)
		}
		log.Println("<<<<==== Received Message : " + receivedMsg)
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

	go serverMsghandler(c, clientName, &counter)

	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])

	dropCounter, resetCounter := 0, 0

	for i := 1; i <= requests; i++ {
		if atomic.LoadInt32(&counter) == 1 {
			// Connection is closed by server sent "Quit Message"
			return
		}
		msgSent := clientName + "- Request-" + strconv.Itoa(i) + "\n"
		_, sendErr := c.Write([]byte(msgSent))
		if sendErr != nil {
			if strings.Contains(sendErr.Error(), util.ErrMsgConnForciblyClosed) {
				storeConnFailure(clientName, remoteAddr, msgSent, sendErr.Error(), i, dropCounter, c, true)
				return
			}
			if strings.Contains(sendErr.Error(), util.ErrMsgConnAborted) {
				storeConnFailure(clientName, remoteAddr, msgSent, sendErr.Error(), i, dropCounter, c, true)
				return
			}
			if sendErr.Error() == util.ErrMsgEOF {
				storeConnFailure(clientName, remoteAddr, msgSent, sendErr.Error(), i, dropCounter, c, true)
				return
			}
			log.Println("#====== Send Error : ", sendErr.Error())
			dropCounter++
			resetCounter++
			log.Println("Packet dropped with ", clientName, " request : ", i)
			if resetCounter == util.MaxDropPackets {
				storeConnFailure(clientName, remoteAddr, msgSent, "connection is broken.", i, dropCounter, c, true)
				return
			}
		} else {
			resetCounter = 0
		}

		log.Println("====>>>> Message Sent : " + msgSent)

		if ctx.Err() != nil {
			log.Println(util.ConnTerminatedSuccessMsg, "Connection terminated by ctrl+c signal : ", c.RemoteAddr().String(), util.ConnTerminatedMsg)
			return
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	c.Close()
}

func startUdpClient(clientName, remoteAddr string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()
	defer c.Close()

	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])

	for i := 1; i <= requests; i++ {

		text := clientName + "-Request-" + strconv.Itoa(i)

		_, err := c.Write([]byte(text))
		if err != nil {
			log.Printf("Write data failed. Client : %s, Error : %v, Message : %s", clientName, err.Error(), text)
		}

		// buffer to get data
		received := make([]byte, 1024)
		_, err = c.Read(received)
		if err != nil {
			log.Printf("Read data failed. Client : %s, Error : %v, Message : %s", clientName, err.Error(), text)
		}

		log.Print("-> Request send : " + text + " - Response received : " + string(received))

		if ctx.Err() != nil {
			storeConnFailure(clientName, remoteAddr, text, "is cancelled by ctl + c.", i, 0, c, true)
			return
		}

		if i == 1 {
			// Just storing the information
			serverInfoMap.Store(clientName, string(received))
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func handleCtrlC(c chan os.Signal, cancel context.CancelFunc) {
	<-c
	// handle ctrl+c event here
	log.Println("#===== Ctrl + C called ... Exiting in few seconds")
	util.NoExitClient = false
	cancel()
}
