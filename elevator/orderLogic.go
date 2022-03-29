package main

import "Driver-go/elevio"

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

func PederSinOrderLogicMain() {
	var myElevator Elevator
	var elevatorPeers []Elevator
	elevatorPeers = append(elevatorPeers, myElevator)
	var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var sysState SystemState = Initialization

	var completedOrders []elevio.ButtonEvent
	var newOrders []elevio.ButtonEvent
	completeOrderChan := make(chan []elevio.ButtonEvent)
	newOrderChan := make(chan elevio.ButtonEvent)

	//hardware
	LocalInit()
	go LocalControl(&myElevator, MasterOrderPanel, completeOrderChan, newOrderChan)

	//network

	id := NetworkConnect()
	elevIndex := id

	msgTx := make(chan NetworkMessage)
	receivedMessages := make(chan NetworkMessage)
	roleChan := make(chan string)

	go RunNetworkInterface(id, msgTx, receivedMessages, roleChan)

	sysState = Slave
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
			index := msg.ID
			if sysState == Master {
				slaveInfo := ExtractSlaveInformation(msg)
				for _, ord := range slaveInfo.CompletedOrders {
					SetOrder(MasterOrderPanel, ord, OT_NoOrder, index)
				}
				for _, ord := range slaveInfo.NewOrders {
					SetOrder(MasterOrderPanel, ord, OT_Order, index)
				}
			} else {
				masterInfo := ExtractMasterInformation(msg)

			}

		//SEND TO NETWORK
		default:
			switch sysState {
			case Master:

			case Slave:

			}
		}
	}
}
