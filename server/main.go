package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"princepereira/TcpClientServer/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

type tcpConnStruct struct {
	conns map[string]net.Conn
	mutex *sync.Mutex
}

type udpConnStruct struct {
	conns map[string]*net.UDPAddr
	mutex *sync.Mutex
}

var tcpConnCache = tcpConnStruct{conns: make(map[string]net.Conn), mutex: &sync.Mutex{}}
var udpConnCache = udpConnStruct{conns: make(map[string]*net.UDPAddr), mutex: &sync.Mutex{}}

var listener net.Listener
var udpListener *net.UDPConn

var quitServer = make(chan bool)

func (c tcpConnStruct) add(remoteAdd string, conn net.Conn) {
	c.mutex.Lock()
	if prevConn, ok := c.conns[remoteAdd]; ok {
		prevConn.Close()
	}
	c.conns[remoteAdd] = conn
	c.mutex.Unlock()
}

func (c tcpConnStruct) remove(remoteAdd string) net.Conn {
	c.mutex.Lock()
	conn := c.conns[remoteAdd]
	delete(c.conns, remoteAdd)
	c.mutex.Unlock()
	return conn
}

func (c udpConnStruct) add(remoteAdd string, conn *net.UDPAddr) {
	c.mutex.Lock()
	if _, ok := c.conns[remoteAdd]; !ok {
		c.conns[remoteAdd] = conn
	}
	c.mutex.Unlock()
}

func (c udpConnStruct) remove(remoteAdd string) *net.UDPAddr {
	c.mutex.Lock()
	conn := c.conns[remoteAdd]
	delete(c.conns, remoteAdd)
	c.mutex.Unlock()
	return conn
}

func killHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Kill handler called... Waiting for %d seconds...\n", util.PrestopWaitTimeout)
	if listener != nil {
		listener.Close()
	}
	time.Sleep(time.Duration(util.PrestopWaitTimeout) * time.Second)
	for remoteAddr, conn := range tcpConnCache.conns {
		for try := 1; try <= 5; try++ {
			_, quitError := conn.Write([]byte(util.QuitMsg))
			log.Println("Quit message send to ", remoteAddr, " QuitError : ", quitError, " Try :", try, " RemoteAddr:", remoteAddr)
			if quitError != nil && util.IsConnClosed(quitError.Error()) {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		tcpConnCache.remove(remoteAddr)
	}
	for remoteAddr, conn := range udpConnCache.conns {
		_, err := udpListener.WriteToUDP([]byte(util.QuitMsg), conn)
		log.Println("Quit message send to ", remoteAddr, " Error : ", err)
		udpConnCache.remove(remoteAddr)
	}
	fmt.Fprintf(w, "All connections are killed\n")
	time.Sleep(15 * time.Second)
	if udpListener != nil {
		udpListener.Close()
	}
	log.Println("All connections are closed ...")
	quitServer <- true
}

func startKillServer() {
	http.HandleFunc("/kill", killHandler)
	log.Println("KillHandler started on port : ", util.KillPort)
	http.ListenAndServe(fmt.Sprintf(":%d", util.KillPort), nil)
}

func main() {

	args, err := util.ValidateArgs()
	if err != nil {
		util.ServerHelp()
		log.Println(err)
		return
	}

	if args[util.AtribHelp] == util.ConstTrue {
		util.ServerHelp()
		return
	}

	if args[util.AtribVersion] == util.ConstTrue {
		fmt.Println(util.Version)
		return
	}

	serverInfo := util.GetIPAddress()
	args[util.AtribServerInfo] = serverInfo

	util.PrintServerBanner(args)

	proto := args[util.AtribProto]
	address := ":" + args[util.AtribPort]
	util.PrestopWaitTimeout, _ = strconv.Atoi(args[util.AtribTimeoutPrestopWait])

	go startKillServer()

	switch proto {
	case util.ConstTCP:
		invokeTcpServer(proto, address, serverInfo)
	case util.ConstUDP:
		invokeUdpServer(proto, address, serverInfo)
	case util.ConstAll:
		go invokeTcpServer(util.ConstTCP, address, serverInfo)
		time.Sleep(3 * time.Second)
		udpPort, _ := strconv.Atoi(args[util.AtribPort])
		udpPort++
		udpAddress := ":" + strconv.Itoa(udpPort)
		invokeUdpServer(util.ConstUDP, udpAddress, serverInfo)
	default:
		log.Fatal("No Proto defined, hence exiting...")
	}

}

func invokeTcpServer(proto, address, serverInfo string) {
	var err error
	listener, err = net.Listen(proto, address)
	if err != nil {
		log.Println("Failed to start server port : ", address)
		log.Println(err)
		return
	}

	log.Println("TCP Server started on port : ", address)

	defer log.Println("TCP Server stopped ...")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("ACCEPT is failed : ", address)
			log.Println(err)
			if strings.Contains(err.Error(), util.ErrMsgListenClosed) {
				<-quitServer
				log.Println("Quiting server")
				return
			}
			continue
		}
		log.Println("TCP Client Connection Established... ", conn.RemoteAddr())
		conn.Write([]byte(">"))
		tcpConnCache.add(conn.RemoteAddr().String(), conn)
		go handleTcpConnection(conn, serverInfo)
	}

}

func handleTcpConnection(conn net.Conn, serverInfo string) {
	defer conn.Close()
	defer tcpConnCache.remove(conn.RemoteAddr().String())
	defer log.Println("TCP connection gracefully closed for client ", conn.RemoteAddr().String())
	s := bufio.NewScanner(conn)
	var msgSent, receivedMsg string
	var sendError error
	for s.Scan() {
		receivedMsg = s.Text()
		log.Print("-> ", string(receivedMsg))
		msgSent = constructServerResp(receivedMsg, serverInfo)
		_, sendError = conn.Write([]byte(msgSent))
		if sendError != nil {
			log.Println("TCP -> Failed to send message to ", conn.RemoteAddr().String(), " Message : ", msgSent, " Error : ", sendError)
		} else {
			log.Println("TCP -> Message sent success to ", conn.RemoteAddr().String(), " Message : ", msgSent)
		}
	}
}

func invokeUdpServer(proto, address, serverInfo string) {

	s, err := net.ResolveUDPAddr(proto, address)
	if err != nil {
		log.Fatalln("Resolve UDP address failed, hence exiting... Error : ", err)
	}

	udpListener, err = net.ListenUDP(proto, s)
	if err != nil {
		log.Fatalln("Listen UDP address failed, hence exiting... Error : ", err)
	}

	log.Println("UDP Server started on port : ", address)

	defer udpListener.Close()
	for i := 1; i <= 4; i++ {
		go udpServers(udpListener, serverInfo, i)
	}
	udpServers(udpListener, serverInfo, 5)
}

func udpServers(listen *net.UDPConn, serverInfo string, index int) {
	var msgSent, receivedMsg string
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := listen.ReadFromUDP(buffer)
		if n <= 0 {
			log.Println("UDP listener ", index, " exitted")
			return
		}
		receivedMsg = string(buffer[0 : n-1])
		log.Print("-> ", receivedMsg)
		udpConnCache.add(remoteAddr.String(), remoteAddr)
		if err != nil {
			log.Println("Error receiving data : ", err)
		}
		msgSent = constructServerResp(receivedMsg, serverInfo)
		_, err = listen.WriteToUDP([]byte(msgSent), remoteAddr)
		if err != nil {
			log.Println("TCP -> Failed to send message to ", remoteAddr.String(), " Message : ", msgSent, " Error : ", err)
		} else {
			log.Println("TCP -> Message sent success to ", remoteAddr.String(), " Message : ", msgSent)
		}
	}
}

func constructServerResp(receivedMsg, serverInfo string) string {
	sentMsg := fmt.Sprintf("Req: %s, Resp: %s\n", receivedMsg, serverInfo)
	return sentMsg
}
