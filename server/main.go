package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"princepereira/TcpClientServer/util"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConnectionMetrics struct {
	IPAddresses map[string]bool
	IPPorts     map[string]bool
}

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
var connectionMetrics = map[string]ConnectionMetrics{
	util.ConstTCP: {IPAddresses: make(map[string]bool), IPPorts: make(map[string]bool)},
	util.ConstUDP: {IPAddresses: make(map[string]bool), IPPorts: make(map[string]bool)},
}

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

func sendFailedStatus(w http.ResponseWriter, probe string) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "Custom 404 for "+probe)
}

func apiListHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("API List handler called ...")
	str := "APIs Supported : \n"
	str = str + "  <IP>:8090/list \n"
	str = str + "  <IP>:8090/kill \n"
	str = str + "  <IP>:8090/healthz \n"
	str = str + "  <IP>:8090/metrics \n"
	str = str + "  <IP>:8090/resetmetrics \n"
	str = str + "  <IP>:8090/readiness \n"
	str = str + "  <IP>:8090/liveness \n"
	str = str + "  <IP>:8090/toggleprobe \n"
	str = str + "  <IP>:8090/failreadinessprobe \n"
	str = str + "  <IP>:8090/passreadinessprobe \n"
	str = str + "  <IP>:2112/metrics \n"
	str = str + "  <IP>:8090/telnet?uri=<ServiceIP:ServicePort> \n"
	fmt.Fprintln(w, str)
}

func resetConnectionMetricsHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Reset Connection Metrics Handler called ...")
	connectionMetrics = map[string]ConnectionMetrics{
		util.ConstTCP: {IPAddresses: make(map[string]bool), IPPorts: make(map[string]bool)},
		util.ConstUDP: {IPAddresses: make(map[string]bool), IPPorts: make(map[string]bool)},
	}
	fmt.Fprintf(w, "connection metrics are reset\n")
}

func readConnectionMetricsHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Read Connection Metrics Handler called ...")
	resp := util.ConnectionMetrics{}
	if tcpConnections, ok := connectionMetrics[util.ConstTCP]; ok {
		for ip := range tcpConnections.IPAddresses {
			resp.TCP.IPAddresses = append(resp.TCP.IPAddresses, ip)
		}
		for ipPort := range tcpConnections.IPPorts {
			resp.TCP.IPPorts = append(resp.TCP.IPPorts, ipPort)
		}
	}
	if udpConnections, ok := connectionMetrics[util.ConstUDP]; ok {
		for ip := range udpConnections.IPAddresses {
			resp.UDP.IPAddresses = append(resp.TCP.IPAddresses, ip)
		}
		for ipPort := range udpConnections.IPPorts {
			resp.UDP.IPPorts = append(resp.TCP.IPPorts, ipPort)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func toggleProbeHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Toggle handler called ...")
	util.FailReadinessProbe = !util.FailReadinessProbe
	log.Printf("Probe flag is toggled to : %v", util.FailReadinessProbe)
	fmt.Fprintf(w, "Probe flag is toggled to : %v", util.FailReadinessProbe)
}

func failReadinessProbeHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("failReadinessProbeHandler called ...")
	if !util.FailReadinessProbe {
		util.FailReadinessProbe = true
	}
	log.Printf("ReadinessProbe will fail : %v", util.FailReadinessProbe)
	fmt.Fprintf(w, "ReadinessProbe will fail : %v", util.FailReadinessProbe)
}

func passReadinessProbeHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("passReadinessProbeHandler called ...")
	if util.FailReadinessProbe {
		util.FailReadinessProbe = false
	}
	log.Printf("ReadinessProbe will fail : %v", util.FailReadinessProbe)
	fmt.Fprintf(w, "ReadinessProbe will fail : %v", util.FailReadinessProbe)
}

func connectBingHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Health outbound handler called ...")
	_, err := http.Get("http://www.bing.com")
	if err != nil {
		log.Printf("Health outbound failed ...")
		sendFailedStatus(w, "www.bing.com failed")
		return
	}
	fmt.Fprintf(w, "Health outbound passed")
	log.Printf("Health outbound passed...")
}

func readinessProbeHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Readiness handler called ...")
	if util.FailReadinessProbe {
		log.Printf("Readiness is set to fail ...")
		sendFailedStatus(w, "Readiness Probe")
		return
	}
	fmt.Fprintf(w, "Readiness probe passed")
	log.Printf("Readiness probe passed...")
}

func livenessProbeHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Liveness handler called ...")
	if util.FailLivenessProbe {
		log.Printf("Liveness is set to fail ...")
		sendFailedStatus(w, "Liveness Probe")
		return
	}
	fmt.Fprintf(w, "Liveness probe passed")
	log.Printf("Liveness probe passed...")
}

func telnetHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Telnet handler called ...")
	uri := req.URL.Query().Get("uri")
	log.Println("Received URI : ", uri)
	for i := 0; i < 3; i++ {
		conn, err := net.DialTimeout("tcp", uri, util.DialTimeout)
		if err == nil && conn != nil {
			log.Printf("Dial to %s is success at try : %d", uri, i)
			fmt.Fprintf(w, "Success")
			return
		}
		time.Sleep(2 * time.Second)
	}
	log.Printf("Dial to %s is failed.", uri)
	w.WriteHeader(http.StatusBadGateway)
	fmt.Fprint(w, "Failed to connect "+uri)
}

func preStopHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("preStopHandler handler called... Waiting for %d seconds...\n", util.PrestopWaitTimeout)
	log.Printf("Faile probe is set to true")

	if listener != nil {
		listener.Close()
	}

	log.Println("Prestop hook waiting at shutdown wait timeout for :", util.PrestopWaitTimeout, " seconds.")
	time.Sleep(time.Duration(util.PrestopWaitTimeout) * time.Second)

	util.FailReadinessProbe = true
	util.FailLivenessProbe = true

	time.Sleep(time.Duration(util.ApplicationWaitTimeout) * time.Second)

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
	log.Println("Prestop hook waiting at application timeout for :", util.ApplicationWaitTimeout, " seconds.")
	time.Sleep(15 * time.Second)
	if udpListener != nil {
		udpListener.Close()
	}
	log.Println("All connections are closed ...")
	quitServer <- true
}

func startHttpHandler() {
	http.HandleFunc("/kill", preStopHandler)
	log.Println("PreStopHandler started on port : ", util.HttpPort)
	http.HandleFunc("/readiness", readinessProbeHandler)
	log.Println("Health outbound Probe started on port : ", util.HttpPort)
	http.HandleFunc("/healthz", connectBingHandler)
	log.Println("Connection metrics started on port : ", util.HttpPort)
	http.HandleFunc("/metrics", readConnectionMetricsHandler)
	log.Println("Reset connection metrics started on port : ", util.HttpPort)
	http.HandleFunc("/resetmetrics", resetConnectionMetricsHandler)
	log.Println("Readiness Probe started on port : ", util.HttpPort)
	http.HandleFunc("/liveness", livenessProbeHandler)
	log.Println("Liveness Probe started on port : ", util.HttpPort)
	http.HandleFunc("/toggleprobe", toggleProbeHandler)
	log.Println("Toggle Probe started on port : ", util.HttpPort)
	http.HandleFunc("/failreadinessprobe", failReadinessProbeHandler)
	log.Println("Fail Readiness Probe started on port : ", util.HttpPort)
	http.HandleFunc("/passreadinessprobe", passReadinessProbeHandler)
	log.Println("Pass Readiness Probe started on port : ", util.HttpPort)
	http.HandleFunc("/telnet", telnetHandler)
	log.Println("Telnet handler started on port : ", util.HttpPort)
	http.HandleFunc("/list", apiListHandler)
	log.Println("API list handler started on port : ", util.HttpPort)
	http.ListenAndServe(fmt.Sprintf(":%d", util.HttpPort), nil)
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
	util.ApplicationWaitTimeout, _ = strconv.Atoi(args[util.AtribTimeoutApplicationWait])

	go startHttpHandler()

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

func addConnectionsToConnectionMetrics(proto string, conn net.Conn) {
	localAddr := conn.LocalAddr().String()
	remoteAddr := conn.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		connectionMetrics[proto].IPAddresses[ip] = true
		connectionMetrics[proto].IPPorts[fmt.Sprintf("%s-%s", remoteAddr, localAddr)] = true
	}
}

func setKeepAlive(c net.Conn, msg string) {

	strSplit := strings.Split(msg, "|")
	if len(strSplit) < 2 {
		return
	}

	tc, ok := c.(*net.TCPConn)
	if !ok {
		return
	}

	strSplit = strings.Split(strSplit[0], ",")
	for _, v := range strSplit {
		kvs := strings.Split(v, ":")
		if len(kvs) < 2 {
			continue
		}
		if strings.Contains(v, "dka") {
			dka, _ := strconv.ParseBool(kvs[1])
			tc.SetKeepAlive(!dka)
			fmt.Println("Setting Disable Keep Alive :", dka)
			if dka {
				return
			}
		}
		if strings.Contains(v, "tka") {
			tka, _ := strconv.Atoi(kvs[1])
			keepAliveTimeOutInt := time.Duration(tka) * time.Millisecond
			tc.SetKeepAlivePeriod(keepAliveTimeOutInt)
			fmt.Println("Setting Keep Alive Timeout :", tka)
		}
	}
}

func handleTcpConnection(conn net.Conn, serverInfo string) {
	// Replacing for now
	hostName, _ := os.Hostname()
	serverInfo = fmt.Sprintf("%s - [ %s -> %s ]", hostName, conn.RemoteAddr().String(), conn.LocalAddr().String())
	defer conn.Close()
	defer tcpConnCache.remove(conn.RemoteAddr().String())
	defer log.Println("TCP connection gracefully closed for client ", conn.RemoteAddr().String())
	s := bufio.NewScanner(conn)
	var msgSent, receivedMsg string
	var sendError error
	firstPacket := true

	for s.Scan() {
		receivedMsg = s.Text()
		log.Print("-> ", string(receivedMsg))
		if firstPacket && strings.Contains(receivedMsg, "dka") {
			addConnectionsToConnectionMetrics(util.ConstTCP, conn)
			setKeepAlive(conn, receivedMsg)
			firstPacket = false
		}
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
	sentMsg := fmt.Sprintf("Req: %s|Resp: %s\n", receivedMsg, serverInfo)
	return sentMsg
}
