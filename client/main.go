package main

import (
	"bufio"
	"context"
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

var failedCons []string
var serverDetails = make(map[string]string)

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

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(chanSignal, cancel)

	address := fmt.Sprintf("%s:%s", args[util.AtribIpAddr], args[util.AtribPort])

	switch proto {
	case util.ConstTCP:
		invokeTcpClient(proto, address, conns, args, wg, ctx)
	case util.ConstUDP:
		invokeUdpClient(proto, address, conns, args, wg, ctx)
	default:
		log.Fatal("No Proto defined, hence exiting...")
	}

	if len(failedCons) != 0 {
		log.Println("\nNo: of failed connections : ", len(failedCons))
		log.Println("\nFailed connections : \n")
		log.Println("=============================\n")
		log.Println(failedCons)
		log.Println("\n=============================\n")
		log.Println("\nClient Details : ", util.GetIPAddress())
	} else {
		log.Println("\n\n######========= ALL CONNECTIONS ARE CLOSED GRACEFULLY =========######\n")
	}
	log.Println("\n\nExiting TCP Client ...")
}

func invokeUdpClient(proto, address string, conns int, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {

		time.Sleep(2 * time.Second)
		clientName := "UdpClient-" + strconv.Itoa(i)
		udpServer, err := net.ResolveUDPAddr(proto, address)
		if err != nil {
			failedCons = append(failedCons, clientName)
			log.Println("#===== Failed to connect to ", address, " by ", clientName, " . Resolve address failed. Error : ", err, " ======#")
			continue
		}

		c, err := net.DialUDP(proto, nil, udpServer)
		if err != nil {
			failedCons = append(failedCons, clientName)
			log.Println("#===== Failed to connect to ", address, " by ", clientName, " . Dial failed. Error : ", err, " ======#")
			continue
		}

		log.Println("#===== Udp Client Connected : ", clientName, " ======#")
		connMap[clientName] = c
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons))

	for clientName, con := range connMap {
		log.Println("#===== Starting ", clientName, " ======#")
		go startUdpClient(clientName, con, args, wg, ctx)
	}

	wg.Wait()
}

func invokeTcpClient(proto, address string, conns int, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	disableKeepAlive, _ := strconv.ParseBool(args[util.AtribDisableKeepAlive])
	keepAliveTimeOut, _ := strconv.Atoi(args[util.AtribTimeoutKeepAlive])

	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {
		clientName := "TcpClient-" + strconv.Itoa(i)
		time.Sleep(2 * time.Second)
		c, err := net.DialTimeout(proto, address, util.DialTimeout)
		if err != nil {
			failedCons = append(failedCons, clientName)
			log.Println("#===== Failed to connect to ", address, " by ", clientName, " ======#")
			log.Println(err)
			continue
		}

		if tc, ok := c.(*net.TCPConn); ok {
			tc.SetKeepAlive(!disableKeepAlive)
			if !disableKeepAlive {
				keepAliveTimeOutInt := time.Duration(keepAliveTimeOut) * time.Millisecond // Negative time will disable keepAlive
				tc.SetKeepAlivePeriod(keepAliveTimeOutInt)
			}
		}

		log.Println("#===== Tcp Client Connected : ", clientName, " ======#")
		connMap[clientName] = c
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons))

	for clientName, con := range connMap {
		log.Println("#===== Starting ", clientName, " ======#")
		go startTcpClient(clientName, con, args, wg, ctx)
	}

	wg.Wait()
}

func exitClient(clientName, reason string, i, packetsDropped int) {
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"
	server := serverDetails[clientName]
	failedCons = append(failedCons, clientName+" Iteration : "+strconv.Itoa(i)+" Server Details : "+server+" Time : "+myTime)
	log.Println(util.ConnTerminatedFailedMsg + clientName + " " + reason + " Exiting... " + myTime + " Packets dropped : " + strconv.Itoa(packetsDropped) + util.ConnTerminatedMsg)
}

func serverMsghandler(conn net.Conn, clientName string, counter *int32) {
	firstMsg := true
	s := bufio.NewScanner(conn)
	for s.Scan() {
		receivedMsg := string(s.Text())
		if firstMsg {
			// Just storing the information
			firstMsg = false
			serverDetails[clientName] = receivedMsg
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

func startTcpClient(clientName string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()
	defer c.Close()

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
		msgSent := clientName + " - Request-" + strconv.Itoa(i) + "\n"
		_, sendErr := c.Write([]byte(msgSent))
		if sendErr != nil {
			if strings.Contains(sendErr.Error(), util.ErrMsgConnForciblyClosed) {
				exitClient(clientName, sendErr.Error(), i, dropCounter)
				return
			}
			if strings.Contains(sendErr.Error(), util.ErrMsgConnAborted) {
				exitClient(clientName, sendErr.Error(), i, dropCounter)
				return
			}
			if sendErr.Error() == util.ErrMsgEOF {
				exitClient(clientName, sendErr.Error(), i, dropCounter)
				return
			}
			log.Println("#====== Send Error : ", sendErr.Error())
			dropCounter++
			resetCounter++
			log.Println("Packet dropped with ", clientName, " request : ", i)
			if resetCounter == util.MaxDropPackets {
				exitClient(clientName, "connection is broken.", i, dropCounter)
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
}

func startUdpClient(clientName string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()
	defer c.Close()

	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])

	for i := 1; i <= requests; i++ {

		text := clientName + " - Request-" + strconv.Itoa(i) + "\n"

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
			exitClient(clientName, "is cancelled by ctl + c.", i, 0)
			return
		}
		if i == 1 {
			// Just storing the information
			serverDetails[clientName] = string(received)
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func handleCtrlC(c chan os.Signal, cancel context.CancelFunc) {
	<-c
	// handle ctrl+c event here
	log.Println("#===== Ctrl + C called ... Exiting in few seconds")
	cancel()
}
