package main

// case init
// case master
// case slave
// receive reports

//FUNC DOSTUFF(CHAN)

type SystemState int

const (
	Initialization SystemState = 0
	Connect                    = 1
	Slave                      = 2
	Master                     = 3
)

var networkPeers []string

var myElevator Elevator
var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

func RunSystemFSM() {

}

func PederSinOrderLogicMain() {
	var sysState SystemState = Initialization
	//hardware
	SlaveFSMinit()
	go SlaveFSM(&myElevator, MasterOrderPanel)

	//network
	var id string
	id = NetworkConnect(id)

	msgTx := make(chan NetworkMessage)
	receivedMessages := make(chan NetworkMessage)
	roleChan := make(chan string)

	sysState = Connect
	go RunNetworkInterface(msgTx, receivedMessages, roleChan)
}
