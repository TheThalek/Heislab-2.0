package main

import (
	"Driver-go/elevio"
	"strconv"
)

type SystemState int

const (
	Initialization SystemState = 0
	Connect                    = 1
	Slave                      = 2
	Master                     = 3
)

func PederSinOrderLogicMain() {
	var myElevator Elevator
	var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var sysState SystemState = Initialization

	var completeOrders []elevio.ButtonEvent
	var newOrders []elevio.ButtonEvent
	completeOrderChan := make(chan []elevio.ButtonEvent)
	newOrderChan := make(chan elevio.ButtonEvent)

	//hardware
	LocalInit()
	go LocalControl(&myElevator, MasterOrderPanel, completeOrderChan, newOrderChan)

	//network
	id := NetworkConnect()
	elevIndex := id
	myElevator.SetIndex(elevIndex)

	var elevatorPeers [NUMBER_OF_ELEVATORS]Elevator
	elevatorPeers[elevIndex] = myElevator

	msgTx := make(chan NetworkMessage)
	receivedMessages := make(chan NetworkMessage)
	roleChan := make(chan string)

	go RunNetworkInterface(id, msgTx, receivedMessages, roleChan)

	sysState = Slave
	for {
		select {
		case cOrds := <-completeOrderChan:
			completeOrders = append(completeOrders, cOrds...)
		case nOrds := <-newOrderChan:
			newOrders = append(newOrders, nOrds)
		case role := <-roleChan:
			if role == "Master" {
				sysState = Master
			} else if role == "Slave" {
				sysState = Slave
			}

		//RECEIVE FROM NETWORK
		case msg := <-receivedMessages:
			peerID, _ := strconv.Atoi(msg.ID)
			if sysState == Master {
				slaveInfo := ExtractSlaveInformation(msg)
				elevatorPeers[peerID] = Elevator{
					direction:    slaveInfo.direction,
					currentFloor: slaveInfo.currentFloor,
					obs:          slaveInfo.obs,
					priOrder:     elevatorPeers[peerID].priOrder,
					index:        peerID,
				}
				for _, ord := range slaveInfo.CompletedOrders {
					SetOrder(&MasterOrderPanel, ord, OT_Completed, peerID)
				}
				for _, ord := range slaveInfo.NewOrders {
					SetOrder(&MasterOrderPanel, ord, OT_Order, peerID)
				}

			} else {
				masterInfo := ExtractMasterInformation(msg, NUMBER_OF_FLOORS, NUMBER_OF_BUTTONS, NUMBER_OF_ELEVATORS)
				MasterOrderPanel = masterInfo.OrderPanel

				var compOrdersUpdate []elevio.ButtonEvent
				for _, ord := range newOrders {
					if GetOrder(MasterOrderPanel, ord, peerID) != OT_Completed {
						compOrdersUpdate = append(compOrdersUpdate, ord)
					}
				}
				var newOrdersUpdate []elevio.ButtonEvent
				for _, ord := range newOrders {
					if GetOrder(MasterOrderPanel, ord, peerID) != OT_Order {
						newOrdersUpdate = append(newOrdersUpdate, ord)
					}
				}
				myElevator.SetPriOrder(masterInfo.Priorities[peerID].order)
			}

		//SEND TO NETWORK
		default:
			elevatorPeers[elevIndex] = myElevator
			switch sysState {
			case Master:
				elevatorPeers = PrioritizeOrders(&MasterOrderPanel, elevatorPeers)
				myElevator.SetPriOrder(elevatorPeers[elevIndex].GetPriOrder())
				priSlice := [NUMBER_OF_ELEVATORS]RemoteOrder{}
				for i := 0; i < len(priSlice); i++ {
					priSlice[i] = RemoteOrder{
						ID:    strconv.Itoa(i),
						order: elevatorPeers[i].GetPriOrder(),
					}
				}
				masterInfo := MasterInformation{
					OrderPanel: MasterOrderPanel,
					Priorities: priSlice,
				}
				msgTx <- NewMasterMessage(strconv.Itoa(id), masterInfo)
			case Slave:
				slaveInfo := SlaveInformation{
					direction:       myElevator.GetDirection(),
					currentFloor:    myElevator.GetCurrentFloor(),
					obs:             myElevator.GetObs(),
					NewOrders:       newOrders,
					CompletedOrders: completeOrders,
				}
				msgTx <- NewSlaveMessage(strconv.Itoa(id), slaveInfo)
			}
		}
	}
}
