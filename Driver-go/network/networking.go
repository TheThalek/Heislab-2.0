package network

import (
	"Driver-go/elevator"
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
const PERIOD = 1 * time.Second

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
	OrderPanel [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int
	Priorities [elevator.NUMBER_OF_ELEVATORS]RemoteOrder
}
type SlaveInformation struct {
	direction       elevio.MotorDirection
	currentFloor    int
	obs             bool
	NewOrders       []elevio.ButtonEvent
	CompletedOrders []elevio.ButtonEvent
}

func newRemoteOrder(id string, order elevio.ButtonEvent) RemoteOrder {
	return RemoteOrder{ID: id, order: order}
}
func newNetworkMessage(origin MessageOrigin, id string, content string) NetworkMessage {
	return NetworkMessage{Origin: origin, ID: id, Content: content, MessageString: string(origin) + DELIM + id + DELIM + content}
}
func NewMasterMessage(id string, info MasterInformation) NetworkMessage {
	infoString := fmt.Sprint(info.OrderPanel) + DELIM + fmt.Sprint(info.Priorities)
	return newNetworkMessage(Master, id, infoString)
}
func NewSlaveMessage(id string, info SlaveInformation) NetworkMessage {
	return newNetworkMessage(Slave, id, strconv.Itoa(info.currentFloor)+DELIM+fmt.Sprint(info.direction)+DELIM+strconv.FormatBool(info.obs)+DELIM+fmt.Sprint(info.CompletedOrders)+DELIM+fmt.Sprint(info.NewOrders))
}

func StringToNetworkMsg(msg string) NetworkMessage {
	var netmsg NetworkMessage
	msgSplit := strings.Split(msg, DELIM)
	netmsg = newNetworkMessage(MessageOrigin(msgSplit[0]), msgSplit[1], strings.Join(msgSplit[2:], DELIM))
	return netmsg
}

//for slave
func ExtractMasterInformation(masterMsg NetworkMessage, numFloors int, numButtons int, numElevs int) MasterInformation {
	mSplit := strings.Split(masterMsg.Content, DELIM)
	o := stringToIntArray(mSplit[0], numFloors, numButtons)
	var orders [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int
	for i := 0; i < len(o); i++ {
		for j := 0; j < len(o[0]); j++ {
			orders[i][j] = o[i][j]
		}
	}
	mSplit[1] = strings.Trim(mSplit[1], "[]")
	mSplit[1] = strings.ReplaceAll(mSplit[1], "{", "")
	mSplit[1] = strings.ReplaceAll(mSplit[1], "}", "")
	priStrArray := strings.Split(mSplit[1], " ")
	var pri [elevator.NUMBER_OF_ELEVATORS]RemoteOrder
	for i := 0; i < len(priStrArray); i = i + 3 {
		id := priStrArray[i]
		fl, _ := strconv.Atoi(priStrArray[i+1])
		btn, _ := strconv.Atoi(priStrArray[i+2])
		pri[i/3] = RemoteOrder{ID: id, order: elevio.ButtonEvent{Floor: fl, Button: elevio.ButtonType(btn)}}
	}
	return MasterInformation{OrderPanel: orders, Priorities: pri}
}

//for master
func ExtractSlaveInformation(slaveMsg NetworkMessage) SlaveInformation {
	mSplit := strings.Split(slaveMsg.Content, DELIM)
	fl, _ := strconv.Atoi(mSplit[0])
	dirInt, _ := strconv.Atoi(mSplit[1])
	dir := elevio.MotorDirection(dirInt)
	ob, _ := strconv.ParseBool(mSplit[2])
	mSplit[3] = strings.Trim(mSplit[3], "[{}]")
	mSplit[3] = strings.ReplaceAll(mSplit[3], "} {", " ")
	mSplit[4] = strings.Trim(mSplit[4], "[{}]")
	mSplit[4] = strings.ReplaceAll(mSplit[4], "} {", " ")

	nOrds := []elevio.ButtonEvent{}
	cOrds := []elevio.ButtonEvent{}
	nOrdsStringArray := strings.Split(mSplit[3], " ")
	for i := 0; i < len(nOrdsStringArray); i = i + 2 {
		fl, _ := strconv.Atoi(nOrdsStringArray[i])
		bt_int, _ := strconv.Atoi(nOrdsStringArray[i+1])
		nOrds = append(nOrds, elevio.ButtonEvent{Floor: fl, Button: elevio.ButtonType(bt_int)})
	}
	cOrdsStringArray := strings.Split(mSplit[4], " ")
	for i := 0; i < len(cOrdsStringArray); i = i + 2 {
		fl, _ := strconv.Atoi(cOrdsStringArray[i])
		bt_int, _ := strconv.Atoi(cOrdsStringArray[i+1])
		cOrds = append(cOrds, elevio.ButtonEvent{Floor: fl, Button: elevio.ButtonType(bt_int)})
	}
	fmt.Println(cOrds, nOrds)
	return SlaveInformation{direction: dir, currentFloor: fl, obs: ob, NewOrders: nOrds, CompletedOrders: cOrds}
}

func ReportMasterTimeOut(masterTimeOutChan chan<- string, reset <-chan string) {
	start := time.Now()
	for {
		select {
		case <-reset:
			start = time.Now()

		default:
			if time.Since(start) > time.Duration(2*PERIOD) {
				masterTimeOutChan <- "Timeout"
				start = time.Now()
			}
		}
	}
}

var orderPanel [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int
var priorityOrders [elevator.NUMBER_OF_ELEVATORS]RemoteOrder = [elevator.NUMBER_OF_ELEVATORS]RemoteOrder{newRemoteOrder("peer--1", elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallUp}), newRemoteOrder("2", elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallUp}), newRemoteOrder("3", elevio.ButtonEvent{Floor: 3, Button: elevio.BT_HallUp})}
var nOrders []elevio.ButtonEvent = []elevio.ButtonEvent{
	{Floor: 3, Button: elevio.BT_Cab},
	{Floor: 1, Button: elevio.BT_HallDown},
}
var cOrders []elevio.ButtonEvent = nOrders

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

	mTimeout := make(chan string)
	resetMasterTimeOut := make(chan string)
	go ReportMasterTimeOut(mTimeout, resetMasterTimeOut)

	state := 0 //slave
	go func() {
		for {
			switch state {
			case 0:
				fmt.Println(id, "is Slave:")
				msgTx <- NewSlaveMessage(id, SlaveInformation{direction: elevio.MD_Up, currentFloor: 2, obs: false, NewOrders: nOrders, CompletedOrders: cOrders})

			case 1:
				fmt.Println(id, "is Master:")
				msgTx <- NewMasterMessage(id, MasterInformation{OrderPanel: orderPanel, Priorities: priorityOrders})
			}
			time.Sleep(PERIOD)
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
			if b.Origin == Master {
				resetMasterTimeOut <- "reset"
			}
			if b.ID != id {
				if b.Origin == Master {
					info := ExtractMasterInformation(b, int(elevator.NUMBER_OF_FLOORS), int(elevator.NUMBER_OF_BUTTONS), int(elevator.NUMBER_OF_ELEVATORS))
					fmt.Printf(id, " Received: %#v\n", info.OrderPanel)
				} else {
					info := ExtractSlaveInformation(b)
					fmt.Printf(id, " Received: %#v\n", info.CompletedOrders)
				}
			}
		case <-mTimeout:
			fmt.Println("Master Timeout")
			state = 1
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

	k := 0
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			el, err := strconv.Atoi(numList[k])
			if err != nil {
				el = 0
			}
			A[i][j] = el
			k++
		}
	}
	return A
}
