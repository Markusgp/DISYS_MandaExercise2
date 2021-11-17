package main

import (
	"fmt"
	"net"
	"strings"
)

var ServConn *net.UDPConn //Connection to receive messages from processes.

func main() {
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	if err != nil {
		println("Could not resolve address localhost:10001 in the shared_resource.go")
	}

	ServConn, err := net.ListenUDP("udp", Address)
	if err != nil {
		println("Could not listen at address localhost:10001 in the shared_resource.go")
	}

	defer ServConn.Close()
	buffer := make([]byte, 1024)
	for {

		bufferLength, _, err := ServConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		message := string(buffer[0:bufferLength]) //message = "id,int_logic_clock,messageText"

		streamMessage := strings.Split(message, ",") //streamMessage = ["id","int_logic_clock","messageText"]

		fmt.Printf("\nProcess ID: %s\nLogical Clock: %s\nMessage Text: %s", streamMessage[0], streamMessage[1], streamMessage[2]) //Prints id, logical clock & message to console
	}
}
