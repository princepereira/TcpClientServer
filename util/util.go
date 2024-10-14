package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Version                  = "v16.10.2024"
	MaxDropPackets           = 100
	HttpPort                 = 8090
	HttpPrometheusPort       = 8090
	ErrMsgConnForciblyClosed = "An existing connection was forcibly closed by the remote host"
	ErrMsgConnAborted        = "An established connection was aborted"
	ErrMsgEOF                = "EOF"
	ErrMsgListenClosed       = "use of closed network connection"
	DialTimeout              = 3 * time.Second
	QuitMsg                  = "Quit Connection"
	ConnTerminatedSuccessMsg = "#####====== Connection graceful exit. "
	ConnTerminatedFailedMsg  = "#####====== Connection failed exit. "
	ConnTerminatedMsg        = " #=== Connection terminated."
	ErrorNotImplemented      = "Not Implemented"
)

const (
	AtribHelp                   = "-h"
	AtribIpAddr                 = "-i"
	AtribPort                   = "-p"
	AtribSrcPort                = "-sp"
	AtribSrcIP                  = "-si"
	AtribCons                   = "-c"
	AtribReqs                   = "-r"
	AtribDelay                  = "-d"
	AtribProto                  = "-pr"
	AtribDisableKeepAlive       = "-dka"
	AtribEnableMetrics          = "-em"
	AtribTimeoutKeepAlive       = "-tka"
	AtribTimeoutPrestopWait     = "-swt"
	AtribTimeoutApplicationWait = "-awt"
	AtribIterations             = "-it"
	AtribMaxDropThreshold       = "-mdt"
	AtribServerInfo             = "ServerInfo"
	AtribVersion                = "-v"
	AtribOutFile                = "-file"
	AtribTestName               = "-name"
)

var argKeys = map[string]bool{
	AtribHelp:                   true,
	AtribIpAddr:                 true,
	AtribPort:                   true,
	AtribSrcPort:                true,
	AtribSrcIP:                  true,
	AtribCons:                   true,
	AtribReqs:                   true,
	AtribDelay:                  true,
	AtribProto:                  true,
	AtribDisableKeepAlive:       true,
	AtribEnableMetrics:          true,
	AtribTimeoutKeepAlive:       true,
	AtribIterations:             true,
	AtribTimeoutPrestopWait:     true,
	AtribTimeoutApplicationWait: true,
	AtribVersion:                true,
	AtribMaxDropThreshold:       true,
	AtribOutFile:                true,
	AtribTestName:               true,
}

var (
	PrestopWaitTimeout     = 15 // In seconds
	ApplicationWaitTimeout = 15 // In seconds
	NoExitClient           = true
	MaxDropThreshold       = 10
	FailReadinessProbe     = false
	FailLivenessProbe      = false
)

const (
	ConstEmpty                    = ""
	ConstFalse                    = "false"
	ConstTrue                     = "true"
	ConstTCP                      = "tcp"
	ConstUDP                      = "udp"
	ConstAll                      = "all"
	DefaultProto                  = ConstTCP
	DefaultDisableKeepAlive       = ConstFalse
	DefaultDisableMetrics         = ConstFalse
	DefaultTimeoutKeepAlive       = "15000"
	DefaultTimeoutPrestopWait     = "5"
	DefaultTimeoutApplicationWait = "15"
	DefaultIterations             = "1"
	DefaultSrcPort                = "-1"
)

type ConnInfo struct {
	ClientName    string `json:",omitempty"`
	LocalAddess   string `json:",omitempty"`
	RemoteAddress string `json:",omitempty"`
	ServerInfo    string `json:",omitempty"`
	RequestInfo   string `json:",omitempty"`
	FailedReason  string `json:",omitempty"`
	FailedTime    string `json:",omitempty"`
}

type Metrics struct {
	IPAddresses []string `json:"ip_addresses,omitempty"`
	IPPorts     []string `json:"connections,omitempty"`
}

type ConnectionMetrics struct {
	TCP Metrics `json:"tcp,omitempty"`
	UDP Metrics `json:"udp,omitempty"`
}

type ConnectionResult struct {
	TestcaseName          string `json:"testcase_name"`
	TotalConnections      int    `json:"total_connections"`
	FailedConnections     int    `json:"failed_connections"`
	SuccessfulConnections int    `json:"successful_connections"`
	Iteration             int    `json:"iteration"`
}

func PrintServerBanner(config map[string]string) {
	log.Println(" ")
	log.Println("#===========================================#")
	log.Println("#         Title            : Network Monitor Server          ")
	log.Printf("#         Version          : %s          \n", Version)
	log.Printf("#         Proto            : %s        \n", config[AtribProto])
	if config[AtribProto] == ConstAll {
		log.Printf("#         TcpPort          : %s               \n", config[AtribPort])
		port, _ := strconv.Atoi(config[AtribPort])
		port++
		log.Printf("#         UdpPort          : %s               \n", strconv.Itoa(port))
	} else {
		log.Printf("#         Port             : %s               \n", config[AtribPort])
	}
	log.Printf("#         Server           : %s               \n", config[AtribServerInfo])
	log.Printf("#         ShutdownWait     : %s               \n", config[AtribTimeoutPrestopWait])
	log.Printf("#         ApplicationWait  : %s               \n", config[AtribTimeoutApplicationWait])
	log.Println("#===========================================#")
	log.Println(" ")
}

func PrintClientBanner(config map[string]string) {
	log.Println(" ")
	log.Println("#===========================================#")
	log.Println("#         Title            : Network Monitor Client          ")
	log.Printf("#         Version          : %s          \n", Version)
	if strings.Contains(config[AtribIpAddr], ":") {
		// printing ipv6 host
		log.Printf("#         Target Host      : [%s]:%s               \n", config[AtribIpAddr], config[AtribPort])
	} else {
		// printing ipv4 host
		log.Printf("#         Target Host      : %s:%s               \n", config[AtribIpAddr], config[AtribPort])
	}
	if p, ok := config[AtribSrcPort]; ok && p != "-1" {
		if strings.Contains(config[AtribSrcIP], ":") {
			log.Printf("#         Source Host      : [%s]:%s                  \n", config[AtribSrcIP], p)
		} else {
			log.Printf("#         Source Host      : %s:%s                  \n", config[AtribSrcIP], p)
		}
	}
	log.Printf("#         Connections      : %s                  \n", config[AtribCons])
	log.Printf("#         Reqs/Cons        : %s                  \n", config[AtribReqs])
	log.Printf("#         Proto            : %s                  \n", config[AtribProto])
	log.Printf("#         Iterations       : %s                  \n", config[AtribIterations])
	log.Printf("#         MaxDropThreshold : %s                  \n", config[AtribMaxDropThreshold])
	if config[AtribProto] == ConstTCP {
		log.Printf("#         DisableKeepAlive : %s                  \n", config[AtribDisableKeepAlive])
		log.Printf("#         TimeoutKeepAlive : %s                  \n", config[AtribTimeoutKeepAlive])
	}
	log.Printf("#         MetricsEnabled   : %s                  \n", config[AtribEnableMetrics])
	if config[AtribOutFile] != ConstEmpty {
		log.Printf("#         ResultFile       : %s                  \n", config[AtribOutFile])
	}
	if config[AtribTestName] != ConstEmpty {
		log.Printf("#         TestcaseName     : %s                  \n", config[AtribTestName])
	}
	log.Println("#===========================================#")
	log.Println(" ")
}

func SetMaxDropThreshold(val string) {
	if val == "" {
		return
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		log.Println("MaxDropThreshold value passed is not an integer, so setting back to default value : ", MaxDropThreshold)
		return
	}
	MaxDropThreshold = intVal
}

func ValidateArgs() (map[string]string, error) {

	var args = make(map[string]string)

	args[AtribProto] = DefaultProto
	args[AtribSrcPort] = DefaultSrcPort
	args[AtribDisableKeepAlive] = DefaultDisableKeepAlive
	args[AtribEnableMetrics] = DefaultDisableMetrics
	args[AtribTimeoutKeepAlive] = DefaultTimeoutKeepAlive
	args[AtribTimeoutPrestopWait] = DefaultTimeoutPrestopWait
	args[AtribTimeoutApplicationWait] = DefaultTimeoutApplicationWait
	args[AtribIterations] = DefaultIterations
	args[AtribMaxDropThreshold] = strconv.Itoa(MaxDropThreshold)
	args[AtribOutFile] = ConstEmpty
	args[AtribTestName] = ConstEmpty

	for i := 1; i < len(os.Args); i++ {

		if i%2 == 1 {

			// Validating argument
			attrib := strings.ToLower(os.Args[i])
			if _, ok := argKeys[attrib]; !ok {
				return nil, fmt.Errorf("unsupported attribute : %s, supported format : %v", attrib, argKeys)
			}

			if i == 1 && attrib == AtribHelp {
				args[attrib] = "true"
				return args, nil
			}

			if i == 1 && attrib == AtribVersion {
				args[attrib] = "true"
				return args, nil
			}

		} else {

			// TODO: Validate Values
			// Assigning values
			attrib := strings.ToLower(os.Args[i-1])
			args[attrib] = strings.ToLower(os.Args[i])
		}
	}

	return args, nil
}

func isValidProto(proto string) bool {
	if proto == ConstTCP {
		return true
	}
	if proto == ConstUDP {
		return true
	}
	return proto == ConstAll
}

func isValidBool(val string) bool {
	if val == ConstTrue {
		return true
	}
	return val == ConstFalse
}

func ValidateValues(cs string, args map[string]string) error {
	if _, err := strconv.Atoi(args[AtribPort]); err != nil {
		return fmt.Errorf("port (%s) should be a number. Error : %v", args[AtribPort], err)
	}
	if _, err := strconv.Atoi(args[AtribCons]); err != nil {
		return fmt.Errorf("connections (%s) should be a number. Error : %v", args[AtribCons], err)
	}
	if _, err := strconv.Atoi(args[AtribReqs]); err != nil {
		return fmt.Errorf("request (%s) should be a number. Error : %v", args[AtribReqs], err)
	}
	if _, err := strconv.Atoi(args[AtribDelay]); err != nil {
		return fmt.Errorf("delay (%s) should be a number. Error : %v", args[AtribDelay], err)
	}
	if val := args[AtribProto]; !isValidProto(val) {
		return fmt.Errorf("proto (%s) should be TCP/UDP", args[AtribProto])
	}
	if val := args[AtribDisableKeepAlive]; !isValidBool(val) {
		return fmt.Errorf("DisableKeepAlive (%s) should be True/False", args[AtribDisableKeepAlive])
	}
	if _, err := strconv.Atoi(args[AtribTimeoutKeepAlive]); err != nil {
		return fmt.Errorf("TimeoutKeepAlive (%s) should be a number. Error : %v", args[AtribTimeoutKeepAlive], err)
	}
	if val := args[AtribEnableMetrics]; !isValidBool(val) {
		return fmt.Errorf("AtribEnableMetrics (%s) should be True/False", args[AtribEnableMetrics])
	}
	if _, err := strconv.Atoi(args[AtribTimeoutPrestopWait]); err != nil {
		return fmt.Errorf("TimeoutPrestopWait (%s) should be a number. Error : %v", args[AtribTimeoutPrestopWait], err)
	}
	if _, err := strconv.Atoi(args[AtribTimeoutApplicationWait]); err != nil {
		return fmt.Errorf("TimeoutApplicationWait (%s) should be a number. Error : %v", args[AtribTimeoutApplicationWait], err)
	}
	if _, err := strconv.Atoi(args[AtribMaxDropThreshold]); err != nil {
		return fmt.Errorf("MaxDropThreshold (%s) should be a number. Error : %v", args[AtribMaxDropThreshold], err)
	}
	if _, err := strconv.Atoi(args[AtribIterations]); err != nil {
		return fmt.Errorf("iterations (%s) should be a number. Error : %v", args[AtribIterations], err)
	}
	srcPort, err := strconv.Atoi(args[AtribSrcPort])
	if err != nil {
		return fmt.Errorf("source port (%s) should be a number. Error : %v", args[AtribSrcPort], err)
	}
	if _, ok := args[AtribSrcIP]; srcPort > 0 && !ok {
		return fmt.Errorf("source ip needs to be specified if source port is specified")
	}
	return nil
}

func ClientHelp() {
	str := "\n#==============================#\n\n"
	str = str + "Format : .\\client.exe -i <IP> -p <Port> -c <Number of Connections> -r <Number of Requests/Connection> -d <Delay (in ms) between each request> \n"
	str = str + "\nEg : .\\client.exe -i 127.0.0.1 -p 4444 -c 1 -r 10000 -d 1 \n"
	str = str + "\nParameters (Optional, Mandatory*): \n\n"
	str = str + "  -i    : (*) IPv4/IPv6 Address of the server \n"
	str = str + "  -p    : (*) Port number of the server \n"
	str = str + "  -c    : (*) Number of clients/threads/connections \n"
	str = str + "  -r    : (*) Number of requests per connection \n"
	str = str + "  -d    : (*) Delay/Sleep/Time between each request for a single connection (in milliseconds) \n"
	str = str + "  -sp   :     Source port to be chosen for client connections. Mandatory to provide Source Ip (-si) as well if this option is specified \n"
	str = str + "  -si   :     Source IP to be chosen for client connections. Only valid if Source Port is specified \n"
	str = str + "  -it   :     Number of iterations. Default: 1 \n"
	str = str + "  -pr   :     Proto used. Options: TCP/UDP. Default: TCP \n"
	str = str + "  -mdt  :     MaxDropThreshold. Max time wait before consecutive drops \n"
	str = str + "  -dka  :     Disable KeepAlive. Options: True/False. Default: False \n"
	str = str + "  -tka  :     KeepAlive Time in milliseconds. Default: 15 seconds \n"
	str = str + "  -em   :     Enable prometheus metrics. Default: False \n"
	str = str + "  -file :     Set the file name to store the result \n"
	str = str + "  -name :     Set the name of the testcase \n"
	str = str + "\n#==============================#\n"
	log.Println(str)
}

func ServerHelp() {
	str := "\n#==============================#\n\n"
	str = str + "Format : .\\server.exe -p <Port> \n"
	str = str + "\nEg : .\\server.exe -p 4444 \n"
	str = str + "\nParameters (Optional, Mandatory*): \n\n"
	str = str + "   -p   : (*) Port number of the server \n"
	str = str + "   -pr  :     Proto used. Options: TCP/UDP/All. Default: TCP \n"
	str = str + "   -swt :     ShutdownWaitTimeout before prestop action in seconds. Default 15 seconds\n"
	str = str + "   -awt :     ApplicationWaitTimeout for cleanup wait in seconds. Default 15 seconds\n"
	str = str + "\n#==============================#\n"
	log.Println(str)
}

func GetClientName(proto string, iter, conn int) string {
	return fmt.Sprintf("Iter-%d-%sClient-%d", iter, strings.ToUpper(proto), conn)
}

func GetIPAddress() string {

	name, err := os.Hostname()
	if err != nil {
		log.Printf("Oops: %v\n", err)
		return ""
	}

	hostDetails := name

	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Printf("Oops: %v\n", err)
		return ""
	}

	for _, a := range addrs {
		hostDetails = hostDetails + " - " + a
	}

	return hostDetails
}

func IsConnClosed(errMsg string) bool {
	if strings.Contains(errMsg, ErrMsgConnForciblyClosed) {
		return true
	}
	if strings.Contains(errMsg, ErrMsgConnAborted) {
		return true
	}
	if errMsg == ErrMsgEOF {
		return true
	}
	if strings.Contains(errMsg, ErrMsgListenClosed) {
		return true
	}
	return false
}

// SaveResultToFile : Save the result to a file
func SaveResultToFile(total, succeeded, failed, iteration int, testcaseName, resultFilePath string) {
	log.Printf("Saving the result of test %s to the file : %s\n", testcaseName, resultFilePath)
	// Save the result to a file
	connresult := ConnectionResult{
		TestcaseName:          testcaseName,
		TotalConnections:      total,
		SuccessfulConnections: succeeded,
		FailedConnections:     failed,
		Iteration:             iteration,
	}
	jsonData, err := json.MarshalIndent(connresult, "", "  ")
	if err != nil {
		log.Println("Error in marshalling the result : ", err)
		return
	}
	err = os.WriteFile(resultFilePath, jsonData, 0644)
	if err != nil {
		log.Println("Error in saving the result : ", err)
		return
	}
	log.Printf("Result of test %s saved successfully to the file : %s\n", testcaseName, resultFilePath)
}

func RemoveResultFile(resultFilePath string) {
	_, err := os.Stat(resultFilePath)
	if os.IsNotExist(err) {
		return
	}
	err = os.Remove(resultFilePath)
	if err != nil {
		log.Println("Error in removing the file : ", err)
	}
}
