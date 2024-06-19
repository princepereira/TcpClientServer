package main

import (
	"princepereira/TcpClientServer/util"
)

func main() {
	go util.StartHttpServer()
	util.StartPrometheus()
}
