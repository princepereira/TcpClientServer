package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide port number")
		fmt.Println("Eg: server 889")
		return
	}

	PORT := ":" + arguments[1]
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
			c.Write([]byte(myTime))
		}
	}
}
