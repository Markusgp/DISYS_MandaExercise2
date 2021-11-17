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

//Ports and servers
var myPort string                    //client own serverport
var numOfServers int                 //Number of servers
var clientConnections []*net.UDPConn //Vector with connections to servers of the other clients
var serverConnection *net.UDPConn    //Connection to own server, (where receive messages.)
var sharedResource *net.UDPConn      //Connection to shared resource

//Identifiers
var id int              //ID of this client
var myLogicalClock int  //Logical clock of this client
var IAmExecutingCS bool //Executing CS?
var iAmWaiting bool     //Waiting for CS?
var requestLC int       //requested logical clock

//replies and requests
var queuedRequests []int    //Queued requests
var repliesRecieved []int   //list of recieved replies
var receivedAllReplies bool //Have I received all replies?

//-----------------------------------------------------------------------
//Set-up and main methods
//-----------------------------------------------------------------------
func main() {
	InitConnections()
	IAmExecutingCS = false
	iAmWaiting = false

	defer serverConnection.Close()
	for i := 0; i < numOfServers; i++ {
		defer clientConnections[i].Close()
	}

	ch := make(chan string)

	go ReadTerminalInput(ch)

	for {
		go ReceiveRequests() //Server

		select {
		case request, valid := <-ch:
			if valid {
				compare, _ := strconv.Atoi(request)
				if compare != id && request == "request" {
					// See if it's in CS or waiting
					if IAmExecutingCS || iAmWaiting {
						fmt.Println("request ignored!")
					} else {
						fmt.Printf("\n --------------- NEW REQUEST ---------------")
						fmt.Printf("\nRequesting access w <ID:%d, LC:%d\n>", id, myLogicalClock)
						textMessage := "Potential text here"
						requestLC = myLogicalClock
						go Ricart_Agrawala(requestLC, textMessage)
					}
				}
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			time.Sleep(time.Second * 1)
		}
		time.Sleep(time.Second * 5)
	}
}

func InitConnections() {
	id, _ = strconv.Atoi(os.Args[1]) //ID becomes the first argument of OS
	myPort = os.Args[id+1]           //Port nr. is the argument of the id +1
	numOfServers = len(os.Args) - 2  //numOfServers is the number of servers assigned in prompt

	connections := make([]*net.UDPConn, numOfServers, numOfServers)

	for i := 0; i < numOfServers; i++ { //Assign all the servers
		port := os.Args[i+2]
		ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+string(port))
		PrintError(err)
		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		PrintError(err)
		connections[i], err = net.DialUDP("udp", LocalAddr, ServerAddr)
		PrintError(err)
	}
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	PrintError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	PrintError(err)
	sharedResource, err = net.DialUDP("udp", LocalAddr, ServerAddr)
	PrintError(err)

	clientConnections = connections

	ServerAddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:"+myPort)
	if err != nil {
		fmt.Println("Failed to resolve port at : " + myPort)
	}
	serverConnection, err = net.ListenUDP("udp", ServerAddr) /* Now listen at selected port */
	if err != nil {
		fmt.Println("Failed to listen at : " + myPort)
	}

	myLogicalClock = 0 //Init the client with a logical clock of 0
}

//-----------------------------------------------------------------------
//Receiving requests from other clients && its helper functions
//-----------------------------------------------------------------------
func ReceiveRequests() {
	buffer := make([]byte, 1024)
	for {

		lengthOfBuffer, _, err := serverConnection.ReadFromUDP(buffer)

		contents := string(buffer[0:lengthOfBuffer]) //contents = "ID,LC,Type"

		streamMessage := strings.Split(contents, ",")     //streamMSG = ["id","LC","Type"]
		receivedID, err := strconv.Atoi(streamMessage[0]) //Received ID
		receivedLC, err := strconv.Atoi(streamMessage[1]) //Received LogicalClock

		//do not send to self
		if receivedID == id {
			continue
		}
		if streamMessage[2] == "request" {

			//I am executingCS or i am waiting (and priority is mine) -> queue.
			if IAmExecutingCS || (iAmWaiting && CheckPriority(receivedID, receivedLC)) {

				if IAmExecutingCS {
					fmt.Printf("\nQueued id: %d w. clock: %d, I am executing CS", receivedID, receivedLC)
					fmt.Printf("\n -> MyID: %d, MyClock: %d", id, myLogicalClock)
					UpdateClock(receivedLC, receivedID, streamMessage[2])
					queuedRequests = append(queuedRequests, receivedID)
					break

				} else if iAmWaiting && CheckPriority(receivedID, receivedLC) {
					fmt.Printf("\nQueued id: %d w. clock: %d, I am waiting with priority", receivedID, receivedLC)
					fmt.Printf("\n -> MyID: %d, MyClock: %d", id, myLogicalClock)
					UpdateClock(receivedLC, receivedID, streamMessage[2])
					queuedRequests = append(queuedRequests, receivedID)
					break
				}
				//send reply instantly
			} else {
				fmt.Printf("\nReplying to id: %d w. clock: %d", receivedID, receivedLC)
				fmt.Printf("\n -> MyID: %d, MyClock: %d", id, myLogicalClock)
				UpdateClock(receivedLC, receivedID, streamMessage[2])
				ReplyToForeign(receivedID)
			}
			//If message is reply
		} else if streamMessage[2] == "reply" {
			if !CheckRepliers(receivedID) { //If reply has not been received, add it
				UpdateClock(receivedLC, receivedID, streamMessage[2])
				repliesRecieved = append(repliesRecieved, receivedID)
			}

			if len(repliesRecieved) >= numOfServers-1 { //If all replies have been received
				fmt.Print("\n----- All replies have been recieved -----")
				receivedAllReplies = true
			}
		} else {
			fmt.Print("\n ERROR -> Cannot understand the message sent")
		}

		if err != nil {
			fmt.Println("ERROR -> ", err)
		}
	}
}

func ReplyToForeign(foreignID int) {
	strConvID := strconv.Itoa(id)
	strConvLC := strconv.Itoa(myLogicalClock)

	message := strConvID + "," + strConvLC + ",reply" //concatenate Message

	index := foreignID - 1
	buffer := []byte(message) //create buffered message

	_, err := clientConnections[index].Write(buffer) //reply to foreign client
	if err != nil {
		fmt.Println(message, err)
	}
}

func ReplyToAnyQueuedRequests() {
	strConvLC := strconv.Itoa(myLogicalClock)
	strConvID := strconv.Itoa(id)

	message := strConvID + "," + strConvLC + ",reply"

	buffer := []byte(message)

	for _, id := range queuedRequests { //Reply to all queued clients
		index := id - 1
		_, err := clientConnections[index].Write(buffer)
		if err != nil {
			fmt.Println(message, err)
		}
	}
}

func CheckRepliers(foreignID int) bool {
	for _, i := range repliesRecieved {
		if i == foreignID {
			return true
		}
	}
	return false
}

func UpdateClock(receivedLC int, receivedID int, msgType string) {
	myLogicalClock = MaxInt(myLogicalClock, receivedLC) + 1 //Update clients' logical clock.
	fmt.Printf("\n__ Received from %d -> updated myClock to : %d, msgtype: "+msgType, receivedID, myLogicalClock)
}

func CheckPriority(foreignID int, foreignClock int) bool {
	if requestLC < foreignClock {
		return true
	} else if foreignClock > requestLC {
		return false
	} else { //If this LC == foreign LC, prioritze lowest ID
		if id < foreignID {
			return true
		} else {
			return false
		}
	}
}

func MaxInt(num1, num2 int) int {
	if num1 > num2 {
		return num1
	}
	return num2
}

//-----------------------------------------------------------------------
//Implementation of algorithm and accessing shared_resource.
//-----------------------------------------------------------------------
func Ricart_Agrawala(requestedLC int, textMessage string) {
	iAmWaiting = true
	RequestAccesToCS(requestedLC)

	fmt.Print("\n----- Waiting for all replies -----")
	for !receivedAllReplies {
	} //Wait until received N-1 replies

	fmt.Print("\n----- Accessing CS ------") //Access the CS
	UseCS(requestedLC, textMessage)

	ReleaseCS()
	fmt.Print("\n----- Released CS ------")
	fmt.Print("\n\n")
}

func RequestAccesToCS(lclock int) {
	strConvLC := strconv.Itoa(lclock)
	strConvID := strconv.Itoa(id)

	message := strConvID + "," + strConvLC + ",request"

	buffer := []byte(message)

	for _, connToClient := range clientConnections { //Multicast request to all n-1 client
		_, err := connToClient.Write(buffer)
		if err != nil {
			fmt.Println(message, err)
		}
	}
}

func UseCS(requestedLC int, textMessage string) {
	IAmExecutingCS = true
	strConvID := strconv.Itoa(id)
	strConvRequestedLC := strconv.Itoa(requestedLC)

	message := strConvID + "," + strConvRequestedLC + "," + textMessage

	buffer := []byte(message)

	_, err := sharedResource.Write(buffer) //send message to sharedResource
	if err != nil {
		fmt.Println(message, err)
	}
	time.Sleep(time.Second * 3)
}

func ReleaseCS() {
	IAmExecutingCS = false
	iAmWaiting = false
	receivedAllReplies = false //release all booleans that locking depends on
	ReplyToAnyQueuedRequests() //reply to any queued requests.
	repliesRecieved = nil      //clear received replies list
}

//-----------------------------------------------------------------------
//General supporting functions
//-----------------------------------------------------------------------
func ReadTerminalInput(ch chan string) {
	reader := bufio.NewReader(os.Stdin)
	for { //Non-blocking routine for listening to terminal input
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
