package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	arguments := os.Args
	if len(arguments) != 3 {
		fmt.Println("Please provide host:port and number of requests.")
		fmt.Println("Eg: client 127.0.0.1:889 10")
		return
	}

	CONNECT := arguments[1]
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println("Failed to connect to ", CONNECT)
		fmt.Println(err)
		return
	}

	requests, err := strconv.Atoi(arguments[2])
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 1; i <= requests; i++ {
		time.Sleep(50 * time.Millisecond)
		text := "Request-" + strconv.Itoa(i)
		fmt.Fprintf(c, text+"\n")
		message, sendErr := bufio.NewReader(c).ReadString('\n')
		if sendErr != nil {
			fmt.Println("Connection is broken. Exiting...")
			return
		}
		fmt.Print("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
