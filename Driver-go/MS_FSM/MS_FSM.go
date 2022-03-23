package MS_FSM

import (
	"Driver-go/elevator"
	"Driver-go/network"
	"Driver-go/network/bcast"
	"Driver-go/network/peers"
)

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

var myElevator elevator.Elevator
var orderPanel [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int

func RunSystemFSM() {
	var sysState SystemState = Initialization
	//hardware

	//var OrderPanel [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int
	//SingleElevatorInit()
	//go RunSingleFSM()

	//network
	id := network.Connect()

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	msgTx := make(chan network.NetworkMessage)
	msgRx := make(chan network.NetworkMessage)

	mTimeout := make(chan string)
	resetMasterTimeOut := make(chan string)
	go network.ReportMasterTimeOut(mTimeout, resetMasterTimeOut)

	sysState = Connect
	for {
		select {
		case p := <-peerUpdateCh:
			networkPeers = p.Peers
			switch sysState {
			case Connect:
				if len(networkPeers) == 1 && elevator.NUMBER_OF_ELEVATORS != 1 {
					id = network.Connect()
					go bcast.Transmitter(16569, msgTx)
					go bcast.Receiver(16569, msgRx)
					go peers.Transmitter(15647, id, peerTxEnable)
					go peers.Receiver(15647, peerUpdateCh)
				}
				sysState = Slave
			case Slave:
				if network.SortPeers(networkPeers)[0] == id {
					sysState = Master
				}
			case Master:
				resetMasterTimeOut <- "Reset"

			}
		case <-msgRx:
			switch sysState {
			case Connect:

			case Slave:

			case Master:
				resetMasterTimeOut <- "Timeout"

			}

		case <-mTimeout:
			resetMasterTimeOut <- "Timeout"
		}
	}
}
