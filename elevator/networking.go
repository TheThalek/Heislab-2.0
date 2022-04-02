package main

import (
	"Driver-go/elevio"
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const DELIM = "//"

const PERIOD = 100 * time.Millisecond

type RemoteOrder struct {
	ID    string
	order elevio.ButtonEvent
}

type MessageOrigin string

const (
	MO_Master MessageOrigin = "MASTER"
	MO_Slave                = "SLAVE"
)

type NetworkMessage struct {
	Origin        MessageOrigin
	ID            string
	Content       string
	MessageString string
}
type MasterInformation struct {
	OrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int
	Priorities [NUMBER_OF_ELEVATORS]RemoteOrder
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
	return newNetworkMessage(MO_Master, id, infoString)
}
func NewSlaveMessage(id string, info SlaveInformation) NetworkMessage {
	return newNetworkMessage(MO_Slave, id, strconv.Itoa(info.currentFloor)+DELIM+fmt.Sprint(info.direction)+DELIM+strconv.FormatBool(info.obs)+DELIM+fmt.Sprint(info.NewOrders)+DELIM+fmt.Sprint(info.CompletedOrders))
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
	var orders [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int
	for i := 0; i < len(o); i++ {
		for j := 0; j < len(o[0]); j++ {
			orders[i][j] = o[i][j]
		}
	}
	mSplit[1] = strings.Trim(mSplit[1], "[]")
	mSplit[1] = strings.ReplaceAll(mSplit[1], "{", "")
	mSplit[1] = strings.ReplaceAll(mSplit[1], "}", "")
	priStrArray := strings.Split(mSplit[1], " ")
	var pri [NUMBER_OF_ELEVATORS]RemoteOrder
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
	fmt.Println("SLAVE MESSAGE", mSplit)
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
	if fmt.Sprint(nOrdsStringArray) != "[]" {
		for i := 0; i < len(nOrdsStringArray); i = i + 2 {
			fl, _ := strconv.Atoi(nOrdsStringArray[i])
			bt_int, _ := strconv.Atoi(nOrdsStringArray[i+1])
			nOrds = append(nOrds, elevio.ButtonEvent{Floor: fl, Button: elevio.ButtonType(bt_int)})
		}
	}
	cOrdsStringArray := strings.Split(mSplit[4], " ")
	if fmt.Sprint(cOrdsStringArray) != "[]" {
		for i := 0; i < len(cOrdsStringArray); i = i + 2 {
			fl, _ := strconv.Atoi(cOrdsStringArray[i])
			bt_int, _ := strconv.Atoi(cOrdsStringArray[i+1])
			cOrds = append(cOrds, elevio.ButtonEvent{Floor: fl, Button: elevio.ButtonType(bt_int)})
		}
	}
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

func SortNetworkPeers(peers []string) []int {

	var sortedPeers []int = []int{}
	for _, s := range peers {
		integer, _ := strconv.Atoi(s)
		sortedPeers = append(sortedPeers, integer)
	}
	return sortedPeers
}

func NetworkConnect() int {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	index, _ := strconv.Atoi(id)
	return index
}

func RunNetworkInterface(id int, msgTx <-chan NetworkMessage, receivedMessages chan<- NetworkMessage, roleChan chan<- string, peerChan chan<- []int) {

	var networkPeers []int
	networkPeers = append(networkPeers, id)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	//Before fucking with the nodes: Transmitter and reciever: 15647, changed it 15659
	go peers.Transmitter(15659, strconv.Itoa(id), peerTxEnable)
	go peers.Receiver(15659, peerUpdateCh)

	msgRx := make(chan NetworkMessage)
	//Before fucking with the nodes: Transmitter and reciever: 16569, changed it 16581
	go bcast.Transmitter(16581, msgTx)
	go bcast.Receiver(16581, msgRx)

	mTimeout := make(chan string)
	resetMasterTimeOut := make(chan string)
	go ReportMasterTimeOut(mTimeout, resetMasterTimeOut)

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			networkPeers = SortNetworkPeers(p.Peers)
			fmt.Println("Peer update: ")
			fmt.Println("  Peers:    \n", networkPeers)
			fmt.Println("  New:      \n", p.New)
			fmt.Println("  Lost:     \n", p.Lost)
			peerChan <- networkPeers
		case a := <-msgRx:
			b := StringToNetworkMsg(a.MessageString)
			if b.Origin == MO_Master && len(networkPeers) > 1 {
				resetMasterTimeOut <- "reset"
			} else if len(networkPeers) == 1 {
				roleChan <- string(MO_Slave)
			}
			receivedMessages <- a
		case <-mTimeout:
			if id == networkPeers[0] && len(networkPeers) > 1 {
				roleChan <- string(MO_Master)
			}
		default:
		}
	}
}

func isInSlice(str string, stringSlice []string) bool {
	for _, s := range stringSlice {
		if s == str {
			return true
		}
	}
	return false
}

func isInSliceInt(i int, intSlice []int) bool {
	for _, j := range intSlice {
		if j == i {
			return true
		}
	}
	return false
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
