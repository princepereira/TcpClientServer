package main

import (
	"bufio"
	"context"
	"fmt"
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
		fmt.Println(err)
		return
	}

	if args[util.AtribHelp] == "true" {
		util.ClientHelp()
		return
	}

	if err := util.ValidateValues("client", args); err != nil {
		util.ClientHelp()
		fmt.Println(err)
		return
	}

	util.PrintClientBanner(args)

	conns, _ := strconv.Atoi(args[util.AtribCons])

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go handleCtrlC(c, cancel)

	CONNECT := fmt.Sprintf("%s:%s", args[util.AtribIpAddr], args[util.AtribPort])
	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {
		time.Sleep(2 * time.Second)

		c, err := net.Dial("tcp", CONNECT)
		clientName := "Client-" + strconv.Itoa(i)
		fmt.Println("#===== Connecting ", clientName, " ======#")

		if err != nil {
			failedCons = append(failedCons, clientName)
			fmt.Println("Failed to connect to ", CONNECT)
			fmt.Println(err)
		} else {
			connMap[clientName] = c
		}
	}

	time.Sleep(3 * time.Second)

	wg.Add(conns - len(failedCons))

	for clientName, con := range connMap {
		fmt.Println("#===== Starting ", clientName, " ======#")
		go startClient(clientName, con, args, wg, ctx)
	}

	wg.Wait()
	fmt.Println("\nNo: of failed connections : ", len(failedCons))
	if len(failedCons) != 0 {
		fmt.Println("\nFailed connections : \n")
		fmt.Println("=============================\n")
		fmt.Println(failedCons)
		fmt.Println("\n=============================\n")
		fmt.Println("\nClient Details : ", util.GetIPAddress(""))
	}
	fmt.Println("\n\nAll the connections are closed. Exiting TCP Client ...")
}

func exitClient(clientName, reason string, i, packetsDropped int) {
	t := time.Now()
	myTime := t.Format(time.RFC3339) + "\n"
	server := serverDetails[clientName]
	failedCons = append(failedCons, clientName+" Iteration : "+strconv.Itoa(i)+" Server Details : "+server+" Time : "+myTime)
	fmt.Println(clientName + " " + reason + " Exiting... " + myTime + " Packets dropped : " + strconv.Itoa(packetsDropped))
}

func startClient(clientName string, c net.Conn, args map[string]string, wg *sync.WaitGroup, ctx context.Context) {

	defer wg.Done()
	defer c.Close()

	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])

	dropCounter, resetCounter := 0, 0

	for i := 1; i <= requests; i++ {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		text := clientName + " - Request-" + strconv.Itoa(i) + "\n"
		fmt.Fprintf(c, text+"\n")
		message, sendErr := bufio.NewReader(c).ReadString('\n')
		if sendErr != nil {
			dropCounter++
			resetCounter++
			fmt.Println("Packet dropped with ", clientName, " request : ", i)
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
		fmt.Print("->: " + text + " - " + message)
	}
}

func handleCtrlC(c chan os.Signal, cancel context.CancelFunc) {
	sig := <-c
	// handle ctrl+c event here
	fmt.Println("Ctl+C called ... Exiting in few seconds")
	fmt.Println("\nsignal: ", sig, "\n\n")
	cancel()
}
