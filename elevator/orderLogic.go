package main

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
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

var myElevator Elevator
var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

func RunSystemFSM() {
	var sysState SystemState = Initialization
	//hardware
	SlaveFSMinit()
	go SlaveFSM(&myElevator, MasterOrderPanel)

	//network
	var id string
	id = NetworkConnect(id)

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	msgTx := make(chan NetworkMessage)
	msgRx := make(chan NetworkMessage)

	go bcast.Transmitter(16569, msgTx)
	go bcast.Receiver(16569, msgRx)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	mTimeout := make(chan string)
	resetMasterTimeOut := make(chan string)
	go ReportMasterTimeOut(mTimeout, resetMasterTimeOut)

	sysState = Connect
	for {
		select {
		case p := <-peerUpdateCh:
			networkPeers = p.Peers
			switch sysState {
			case Connect:
				if len(networkPeers) == 1 && NUMBER_OF_ELEVATORS != 1 {
					id = NetworkConnect()
					go bcast.Transmitter(16569, msgTx)
					go bcast.Receiver(16569, msgRx)
					go peers.Transmitter(15647, id, peerTxEnable)
					go peers.Receiver(15647, peerUpdateCh)
				}
				if id == NetworkSortPeers(networkPeers)[0] {
					sysState = Master
					//msgTx <- network.NewMasterMessage(id,)
				} else {
					sysState = Slave
					//msgTx <- network.NewMasterMessage(id,)
				}

			case Slave:
				if NetworkSortPeers(networkPeers)[0] == id {
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

func orderlogicOrders() {
	//ser bare på tilkobling til orders-modulen
}
