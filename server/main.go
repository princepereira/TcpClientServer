package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"princepereira/TcpClientServer/util"
	"strings"
	"time"
)

func getIPAddress() string {

	client := "Client : "

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return ""
	}

	client = client + name

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return ""
	}

	for _, a := range addrs {
		client = client + " - " + a
	}

	return client
}

func main() {

	args, err := util.ValidateArgs()
	if err != nil {
		util.Help()
		fmt.Println(err)
		return
	}

	if args[util.AtribHelp] == "true" {
		util.Help()
		return
	}

	clientName := getIPAddress()

	args["client"] = clientName
	util.PrintServerBanner(args)

	PORT := ":" + args[util.AtribPort]
	l, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println("Failed to start server port : ", PORT)
		fmt.Println(err)
		return
	}

	fmt.Println("Server started on port : ", PORT)

	defer fmt.Println("Server stopped ...")
	defer l.Close()

	for {

		c, err := l.Accept()
		if err != nil {
			fmt.Println("ACCEPT is failed : ", PORT)
			fmt.Println(err)
			return
		}

		fmt.Println("Client Connection Established...")

		for {

			netData, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				break
			}
			if strings.TrimSpace(string(netData)) == "STOP" {
				fmt.Println("Exiting TCP server!")
				break
			}

			fmt.Print("-> ", string(netData))
			t := time.Now()
			myTime := t.Format(time.RFC3339) + "\n"
			c.Write([]byte(clientName + " \n Time : " + myTime))
		}
	}
}
