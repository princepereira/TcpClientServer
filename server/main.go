package main

import (
	"bufio"
	"log"
	"net"
	"princepereira/TcpClientServer/util"
	"strconv"
	"time"
)

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
	l, err := net.Listen(proto, address)
	if err != nil {
		log.Println("Failed to start server port : ", address)
		log.Println(err)
		return
	}

	log.Println("TCP Server started on port : ", address)

	defer log.Println("Server stopped ...")
	defer l.Close()

	for {

		c, err := l.Accept()
		if err != nil {
			log.Println("ACCEPT is failed : ", address)
			log.Println(err)
			return
		}

		log.Println("TCP Client Connection Established... ", c.RemoteAddr())

		for {
			netData, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				log.Println(err)
				break
			}
			log.Print("-> ", string(netData))
			c.Write(constructServerResp(serverInfo))
		}
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
