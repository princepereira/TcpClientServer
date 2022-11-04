package main

import (
	"bufio"
	"log"
	"net"
	"princepereira/TcpClientServer/util"
	"strings"
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

	PORT := ":" + args[util.AtribPort]
	l, err := net.Listen(proto, PORT)
	if err != nil {
		log.Println("Failed to start server port : ", PORT)
		log.Println(err)
		return
	}

	log.Println("Server started on port : ", PORT)

	defer log.Println("Server stopped ...")
	defer l.Close()

	for {

		c, err := l.Accept()
		if err != nil {
			log.Println("ACCEPT is failed : ", PORT)
			log.Println(err)
			return
		}

		log.Println("Client Connection Established...")

		for {

			netData, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				log.Println(err)
				break
			}
			if strings.TrimSpace(string(netData)) == "STOP" {
				log.Println("Exiting TCP server!")
				break
			}

			log.Print("-> ", string(netData))
			t := time.Now()
			myTime := t.Format(time.RFC3339) + "\n"
			c.Write([]byte("Server : " + serverInfo + " \n Time : " + myTime))
		}
	}
}
