package network

import (
	"Driver-go/elevio"
	"Driver-go/network/bcast"
	"Driver-go/network/localip"
	"Driver-go/network/peers"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const DELIM = "//"

type RemoteOrder struct {
	ID    string
	order elevio.ButtonEvent
}

type MessageOrigin string

const (
	Master MessageOrigin = "MASTER"
	Slave                = "SLAVE"
)

type NetworkMessage struct {
	Origin        MessageOrigin
	ID            string
	Content       string
	MessageString string
}
type MasterInformation struct {
	OrderPanel [][]int
	Priorities []RemoteOrder
}
type SlaveInformation struct {
	CompletedOrders []elevio.ButtonEvent
}

func NewNetworkMessage(origin MessageOrigin, id string, content string) NetworkMessage {
	return NetworkMessage{Origin: origin, ID: id, Content: content, MessageString: string(origin) + DELIM + id + DELIM + content}
}
func StringToNetworkMsg(msg string) NetworkMessage {
	var netmsg NetworkMessage
	msgSplit := strings.Split(msg, DELIM)

	netmsg = NewNetworkMessage(MessageOrigin(msgSplit[0]), msgSplit[1], strings.Join(msgSplit[2:], DELIM))

	return netmsg
}

func ExtractMasterInformation(masterMsg NetworkMessage, numFloors int, numButtons int, numElevs int) {
	strArr := stringToIntArray(masterMsg.Content, numFloors, numButtons)
	orderString := fmt.Sprint(strArr)
	//priString := msgSplit[1]

	fmt.Println("Extract: ", orderString)
	//orderPanel := [][]int(msgSplit[0])
}
func ExtractSlaveInformation(slaveMsg NetworkMessage) {

}
func ReportTimeOut() {

}

var orderPanel [4][3]int
var prioriyOrders [3]int

func PederSinMain() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	msgTx := make(chan NetworkMessage)
	msgRx := make(chan NetworkMessage)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, msgTx)
	go bcast.Receiver(16569, msgRx)

	go func() {
		netMsg := NewNetworkMessage(
			Master,
			id,
			fmt.Sprint(orderPanel)+DELIM+
				fmt.Sprint(prioriyOrders))
		for {
			msgTx <- netMsg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-msgRx:
			b := StringToNetworkMsg(a.MessageString)
			ExtractMasterInformation(b, 4, 3, 1)
			fmt.Printf("Received: %#v\n", b.Content)
		}
	}
}

func stringToIntArray(S string, m int, n int) [][]int {
	A := make([][]int, m)
	for i := range A {
		A[i] = make([]int, n)
	}
	S = strings.ReplaceAll(S, "[", "")
	S = strings.ReplaceAll(S, "]", "")
	numList := strings.Split(S, " ")
	for i := range numList {
		j := i / m
		k := (i - i/m) % n
		el, _ := strconv.Atoi(numList[i])
		if i == 4 {

			el = 1
		}
		fmt.Println(j, k)
		A[i][j] = el
	}
	fmt.Println("to array: ", fmt.Sprint(A))
	return A
}
