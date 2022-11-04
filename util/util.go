package util

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	Version        = "v04.11.2022"
	MaxDropPackets = 100
)

const (
	AtribHelp             = "-h"
	AtribIpAddr           = "-i"
	AtribPort             = "-p"
	AtribCons             = "-c"
	AtribReqs             = "-r"
	AtribDelay            = "-d"
	AtribProto            = "-pr"
	AtribDisableKeepAlive = "-dka"
	AtribTimeoutKeepAlive = "-tka"
	AtribServerInfo       = "ServerInfo"
)

var argKeys = map[string]bool{
	AtribHelp:             true,
	AtribIpAddr:           true,
	AtribPort:             true,
	AtribCons:             true,
	AtribReqs:             true,
	AtribDelay:            true,
	AtribProto:            true,
	AtribDisableKeepAlive: true,
	AtribTimeoutKeepAlive: true,
}

const (
	ConstFalse              = "false"
	ConstTrue               = "true"
	ConstTCP                = "tcp"
	ConstUDP                = "udp"
	ConstAll                = "all"
	DefaultProto            = ConstTCP
	DefaultDisableKeepAlive = ConstFalse
	DefaultTimeoutKeepAlive = "15000"
)

func PrintServerBanner(config map[string]string) {
	log.Println(" ")
	log.Println("#===========================================#")
	log.Println("#         Title       : TCP Server          ")
	log.Printf("#         Version     : %s          \n", Version)
	log.Printf("#         Proto       : %s        \n", config[AtribProto])
	if config[AtribProto] == ConstAll {
		log.Printf("#         TcpPort     : %s               \n", config[AtribPort])
		port, _ := strconv.Atoi(config[AtribPort])
		port++
		log.Printf("#         UdpPort     : %s               \n", strconv.Itoa(port))
	} else {
		log.Printf("#         Port        : %s               \n", config[AtribPort])
	}
	log.Printf("#         Server      : %s               \n", config[AtribServerInfo])
	log.Println("#===========================================#")
	log.Println(" ")
}

func PrintClientBanner(config map[string]string) {
	log.Println(" ")
	log.Println("#===========================================#")
	log.Println("#         Title            : TCP Client          ")
	log.Printf("#         Version          : %s          \n", Version)
	log.Printf("#         Host             : %s:%s               \n", config[AtribIpAddr], config[AtribPort])
	log.Printf("#         Connections      : %s                  \n", config[AtribCons])
	log.Printf("#         Reqs/Cons        : %s                  \n", config[AtribReqs])
	log.Printf("#         Proto            : %s                  \n", config[AtribProto])
	if config[AtribProto] == ConstTCP {
		log.Printf("#         DisableKeepAlive : %s                  \n", config[AtribDisableKeepAlive])
		log.Printf("#         TimeoutKeepAlive : %s                  \n", config[AtribTimeoutKeepAlive])
	}
	log.Println("#===========================================#")
	log.Println(" ")
}

func ValidateArgs() (map[string]string, error) {

	var args = make(map[string]string)

	args[AtribProto] = DefaultProto
	args[AtribDisableKeepAlive] = DefaultDisableKeepAlive
	args[AtribTimeoutKeepAlive] = DefaultTimeoutKeepAlive

	for i := 1; i < len(os.Args); i++ {

		if i%2 == 1 {

			// Validating argument
			attrib := strings.ToLower(os.Args[i])
			if _, ok := argKeys[attrib]; !ok {
				return nil, fmt.Errorf("unsupported attribute : %s, supported format : %v", attrib, argKeys)
			}
			// if _, ok := args[attrib]; ok {
			// 	return nil, fmt.Errorf("repeated attribute : %s, supported format : %v", attrib, argKeys)
			// }

			if i == 1 && attrib == AtribHelp {
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
	if len(args) != 8 {
		if cs == "client" {
			ClientHelp()
		} else {
			ServerHelp()
		}
		return fmt.Errorf("no sufficient args")
	}
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
		return fmt.Errorf("DisableKeepAlive (%s) should be True/False", args[AtribProto])
	}
	if _, err := strconv.Atoi(args[AtribTimeoutKeepAlive]); err != nil {
		return fmt.Errorf("TimeoutKeepAlive (%s) should be a number. Error : %v", args[AtribTimeoutKeepAlive], err)
	}
	return nil
}

func ClientHelp() {
	str := "\n#==============================#\n\n"
	str = str + "Format : .\\client.exe -i <IP> -p <Port> -c <Number of Connections> -r <Number of Requests/Connection> -d <Delay (in ms) between each request> \n"
	str = str + "\nEg : .\\client.exe -i 127.0.0.1 -p 4444 -c 1 -r 10000 -d 1 \n"
	str = str + "\nParameters (Optional, Mandatory*): \n\n"
	str = str + "   -i   : (*) IP Address of the server \n"
	str = str + "   -p   : (*) Port number of the server \n"
	str = str + "   -c   : (*) Number of clients/threads/connections \n"
	str = str + "   -r   : (*) Number of requests per connection \n"
	str = str + "   -d   : (*) Delay/Sleep/Time between each request for a single connection (in milliseconds) \n"
	str = str + "   -pr  :     Proto used. Options: TCP/UDP. Default: TCP \n"
	str = str + "   -dka :     Disable KeepAlive. Options: True/False. Default: False \n"
	str = str + "   -tka :     KeepAlive Time in milliseconds. Default: 15 seconds \n"
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
	str = str + "\n#==============================#\n"
	log.Println(str)
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
