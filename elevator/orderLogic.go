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
		case cOrds := <-completeOrderChan:
			completedOrders = append(completedOrders, cOrds...)
		case nOrds := <-newOrderChan:
			newOrders = append(newOrders, nOrds...)
		case role := <-roleChan:
			if role == MO_Master {
				sysState = Master
			} else if role == MO_Slave {
				sysState = Slave
			}
		case IDs := <-peersIDChan:
			var newElevSlice []Elevator
			for _, elevator := range elevatorPeers {
				if isInSlice(elevator.GetID(), IDs) {
					newElevSlice = append(newElevSlice, elevator)
				}
			}
			elevatorPeers = newElevSlice

		case idx := <-elevIndexChanRx:
			elevIndex = idx

		//RECEIVE FROM NETWORK
		case msg := <-receivedMessages:
			index := 
			if sysState == Master {
				slaveInfo := ExtractSlaveInformation(msg)
				for _, ord := range slaveInfo.CompletedOrders {
					SetOrder(MasterOrderPanel, ord, OT_NoOrder, INDEX)
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
