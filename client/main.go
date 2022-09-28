package main

import (
	"bufio"
	"fmt"
	"net"
	"princepereira/TcpClientServer/util"
	"strconv"
	"sync"
	"time"
)

var failedCons []string
var quit bool

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

	if err := util.ValidateValues(args); err != nil {
		util.Help()
		fmt.Println(err)
		return
	}

	util.PrintClientBanner(args)

	conns, _ := strconv.Atoi(args[util.AtribCons])

	wg := new(sync.WaitGroup)

	// c := make(chan os.Signal)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go handleCtrlC(c)

	CONNECT := fmt.Sprintf("%s:%s", args[util.AtribIpAddr], args[util.AtribPort])
	var connMap = make(map[string]net.Conn)

	// Setting up connections
	for i := 1; i <= conns; i++ {
		time.Sleep(100 * time.Millisecond)

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

	time.Sleep(500 * time.Millisecond)

	wg.Add(conns - len(failedCons))

	for clientName, con := range connMap {
		fmt.Println("#===== Starting ", clientName, " ======#")
		go startClient(clientName, con, args, wg)
	}

	wg.Wait()
	fmt.Println("\nNo: of failed connections : ", len(failedCons))
	if len(failedCons) != 0 {
		fmt.Println("\nFailed connections : ", failedCons)
	}
	fmt.Println("\n\nAll the connections are closed. Exiting TCP Client ...")
}

func startClient(clientName string, c net.Conn, args map[string]string, wg *sync.WaitGroup) {

	defer wg.Done()
	defer c.Close()

	requests, _ := strconv.Atoi(args[util.AtribReqs])
	delay, _ := strconv.Atoi(args[util.AtribDelay])

	for i := 1; i <= requests; i++ {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		text := clientName + " - Request-" + strconv.Itoa(i) + "\n"
		fmt.Fprintf(c, text+"\n")
		message, sendErr := bufio.NewReader(c).ReadString('\n')
		if sendErr != nil {
			failedCons = append(failedCons, clientName+" Iteration : "+text)
			fmt.Println(clientName + " connection is broken. Exiting...")
			return
		}
		fmt.Print("->: " + text + " - " + message)
	}

}

// func handleCtrlC(c chan os.Signal) {
// 	sig := <-c
// 	// handle ctrl+c event here
// 	// for example, close database
// 	fmt.Println("\nsignal: ", sig)
// 	quit = true
// 	// os.Exit(0)
// }
