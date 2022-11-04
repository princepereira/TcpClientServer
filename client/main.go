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
	"sync"
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
	disableKeepAlive, _ := strconv.ParseBool(args[util.AtribDisableKeepAlive])
	keepAliveTimeOut, _ := strconv.Atoi(args[util.AtribTimeoutKeepAlive])

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())

	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(chanSignal, cancel)

	address := fmt.Sprintf("%s:%s", args[util.AtribIpAddr], args[util.AtribPort])
	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {

		time.Sleep(2 * time.Second)
		c, err := net.Dial(proto, address)

		if tc, ok := c.(*net.TCPConn); ok {
			tc.SetKeepAlive(!disableKeepAlive)
			if !disableKeepAlive {
				keepAliveTimeOutInt := time.Duration(keepAliveTimeOut) * time.Millisecond // Negative time will disable keepAlive
				tc.SetKeepAlivePeriod(keepAliveTimeOutInt)
			}
		}

		// c, err := net.Dial("tcp", address)
		clientName := "Client-" + strconv.Itoa(i)

		if err != nil {
			failedCons = append(failedCons, clientName)
			log.Println("#===== Failed to connect to ", address, " by ", clientName, " ======#")
			log.Println(err)
		} else {
			log.Println("#===== Client Connected : ", clientName, " ======#")
			connMap[clientName] = c
		}
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons))

	for clientName, con := range connMap {
		log.Println("#===== Starting ", clientName, " ======#")
		go startClient(clientName, con, args, wg, ctx)
	}

	wg.Wait()

	log.Println("\nNo: of failed connections : ", len(failedCons))
	if len(failedCons) != 0 {
		log.Println("\nFailed connections : \n")
		log.Println("=============================\n")
		log.Println(failedCons)
		log.Println("\n=============================\n")
		log.Println("\nClient Details : ", util.GetIPAddress())
	}
	log.Println("\n\nAll the connections are closed. Exiting TCP Client ...")
}

func exitClient(clientName, reason string, i, packetsDropped int) {
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"
	server := serverDetails[clientName]
	failedCons = append(failedCons, clientName+" Iteration : "+strconv.Itoa(i)+" Server Details : "+server+" Time : "+myTime)
	log.Println(clientName + " " + reason + " Exiting... " + myTime + " Packets dropped : " + strconv.Itoa(packetsDropped))
}

func startClient(clientName string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()
	defer c.Close()

	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])

	dropCounter, resetCounter := 0, 0

	for i := 1; i <= requests; i++ {
		text := clientName + " - Request-" + strconv.Itoa(i) + "\n"
		fmt.Fprintf(c, text+"\n")
		message, sendErr := bufio.NewReader(c).ReadString('\n')
		if sendErr != nil {
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
		if ctx.Err() != nil {
			exitClient(clientName, "is cancelled by ctl + c.", i, dropCounter)
			return
		}
		if i == 1 {
			// Just storing the information
			serverDetails[clientName] = message
		}
		log.Print("->: " + text + " - " + message)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func handleCtrlC(c chan os.Signal, cancel context.CancelFunc) {
	<-c
	// handle ctrl+c event here
	log.Println("#===== Ctl+C called ... Exiting in few seconds")
	cancel()
}
