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

var tcpConnCache = tcpConnStruct{conns: make(map[string]net.Conn), mutex: &sync.Mutex{}}
var listener net.Listener
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

func killHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("Kill handler called...")
	if listener != nil {
		listener.Close()
	}
	for remoteAddr, conn := range tcpConnCache.conns {
		conn.Write([]byte(util.QuitMsg))
		log.Println("Quit message send to ", remoteAddr)
		time.Sleep(500 * time.Millisecond)
		conn.Close()
		log.Println("Connection closed for ", remoteAddr)
		tcpConnCache.remove(remoteAddr)
	}
	log.Println("All connections are closed ...")
	quitServer <- true
	fmt.Fprintf(w, "All connections are killed\n")
}

func startKillServer() {
	http.HandleFunc("/kill", killHandler)
	log.Println("killHandler started on port : ", util.KillPort)
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

	serverInfo := util.GetIPAddress()
	args[util.AtribServerInfo] = serverInfo

	util.PrintServerBanner(args)

	proto := args[util.AtribProto]
	address := ":" + args[util.AtribPort]

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
	for s.Scan() {
		data := s.Text()
		log.Print("-> ", string(data))
		conn.Write(constructServerResp(serverInfo))
	}
}

func invokeUdpServer(proto, address, serverInfo string) {

	s, err := net.ResolveUDPAddr(proto, address)
	if err != nil {
		log.Fatalln("Resolve UDP address failed, hence exiting... Error : ", err)
	}

	connection, err := net.ListenUDP(proto, s)
	if err != nil {
		log.Fatalln("Listen UDP address failed, hence exiting... Error : ", err)
	}

	log.Println("UDP Server started on port : ", address)

	defer connection.Close()
	buffer := make([]byte, 1024)

	for {
		n, addr, err := connection.ReadFromUDP(buffer)
		log.Print("-> ", string(buffer[0:n-1]))
		if err != nil {
			log.Println("Error receiving data : ", err)
		}
		_, err = connection.WriteToUDP(constructServerResp(serverInfo), addr)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func constructServerResp(serverInfo string) []byte {
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"
	serverResp := []byte("Server : " + serverInfo + " \n Time : " + myTime)
	return serverResp
}
