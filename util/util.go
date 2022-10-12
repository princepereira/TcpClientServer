package util

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	Version        = "v28.09.2022"
	MaxDropPackets = 100
)

const (
	AtribHelp   = "-h"
	AtribIpAddr = "-i"
	AtribPort   = "-p"
	AtribCons   = "-c"
	AtribReqs   = "-r"
	AtribDelay  = "-d"
)

var argKeys = map[string]bool{
	AtribHelp:   true,
	AtribIpAddr: true,
	AtribPort:   true,
	AtribCons:   true,
	AtribReqs:   true,
	AtribDelay:  true,
}

func PrintServerBanner(config map[string]string) {
	fmt.Println(" ")
	fmt.Println("#===========================================#")
	fmt.Println("#         Title       : TCP Server          #")
	fmt.Printf("#         Version     : %s         # \n", Version)
	fmt.Printf("#         Port       : %s               #\n", config[AtribPort])
	fmt.Printf("#         Client     : %s               #\n", config["client"])
	fmt.Println("#===========================================#")
	fmt.Println(" ")
}

func PrintClientBanner(config map[string]string) {
	fmt.Println(" ")
	fmt.Println("#===========================================#")
	fmt.Println("#         Title       : TCP Client          #")
	fmt.Printf("#         Version     : %s         # \n", Version)
	fmt.Printf("#         Host        : %s:%s               #\n", config[AtribIpAddr], config[AtribPort])
	fmt.Printf("#         Connections : %s                  #\n", config[AtribCons])
	fmt.Printf("#         Reqs/Cons   : %s                  #\n", config[AtribReqs])
	fmt.Println("#===========================================#")
	fmt.Println(" ")
}

func ValidateArgs() (map[string]string, error) {

	var args = make(map[string]string)

	for i := 1; i < len(os.Args); i++ {

		if i%2 == 1 {

			// Validating argument
			attrib := strings.ToLower(os.Args[i])
			if _, ok := argKeys[attrib]; !ok {
				return nil, fmt.Errorf("unsupported attribute : %s, supported format : %v", attrib, argKeys)
			}
			if _, ok := args[attrib]; ok {
				return nil, fmt.Errorf("repeated attribute : %s, supported format : %v", attrib, argKeys)
			}

			if i == 1 && attrib == AtribHelp {
				args[attrib] = "true"
				return args, nil
			}

		} else {

			// TODO: Validate Values
			// Assigning values
			attrib := strings.ToLower(os.Args[i-1])
			args[attrib] = os.Args[i]
		}
	}

	return args, nil
}

func ValidateValues(cs string, args map[string]string) error {
	if len(args) != 5 {
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
	return nil
}

func ClientHelp() {
	str := "\n#==============================#\n\n"
	str = str + "Format : .\\client.exe -i <IP> -p <Port> -c <Number of Connections> -r <Number of Requests/Connection> -d <Delay (in ms) between each request> \n"
	str = str + "Eg : .\\client.exe -i 127.0.0.1 -p 4444 -c 10 -r 10 -d 50 \n"
	str = str + "\n#==============================#\n"
	fmt.Println(str)
}

func ServerHelp() {
	str := "\n#==============================#\n\n"
	str = str + "Format : .\\server.exe -p <Port> \n"
	str = str + "Server Eg : .\\server.exe -p 4444 \n"
	str = str + "\n#==============================#\n"
	fmt.Println(str)
}

func GetIPAddress(str string) string {

	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return ""
	}

	client := str + name

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
