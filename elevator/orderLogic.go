package main

import (
	"Driver-go/elevio"
	"fmt"
	"strconv"
	"time"
)

type SystemState int

const (
	Initialization SystemState = 0
	Connect                    = 1
	Slave                      = 2
	Master                     = 3
)

func PederSinOrderLogicMain() {
	var myElevator Elevator = NewElevator()
	var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var sysState SystemState = Initialization

	var completeOrders []elevio.ButtonEvent
	var newOrders []elevio.ButtonEvent
	completeOrderChan := make(chan []elevio.ButtonEvent)
	newOrderChan := make(chan elevio.ButtonEvent)

	//hardware
	LocalInit()
	go LocalControl(&myElevator, &MasterOrderPanel, completeOrderChan, newOrderChan)

	//network
	id := NetworkConnect()
	elevIndex := id
	myElevator.SetIndex(elevIndex)

	var elevatorPeers [NUMBER_OF_ELEVATORS]*Elevator
	for i := 0; i < len(elevatorPeers); i++ {
		nElev := NewElevator()
		nElev.SetIndex(i)
		elevatorPeers[i] = &nElev
	}

	elevatorPeers[elevIndex] = &myElevator

	msgTx := make(chan NetworkMessage)
	receivedMessages := make(chan NetworkMessage)
	roleChan := make(chan string)
	peerChan := make(chan []int)

	go RunNetworkInterface(id, msgTx, receivedMessages, roleChan, peerChan)

	sysState = Slave
	for {
		select {
		case cOrds := <-completeOrderChan:
			completeOrders = append(completeOrders, cOrds...)
		case nOrds := <-newOrderChan:
			newOrders = append(newOrders, nOrds)
			//fmt.Println(newOrders)
		case role := <-roleChan:
			if role == string(MO_Master) {
				sysState = Master
				//fmt.Println("MY role: ", role)
			} else if role == string(MO_Slave) {
				sysState = Slave
			}

		case onlinePeers := <-peerChan:
			for i := 0; i < NUMBER_OF_ELEVATORS; i++ {
				if isInSliceInt(i, onlinePeers) {
					elevatorPeers[i].SetOnline(true)
				} else {
					elevatorPeers[i].SetOnline(false)
				}
			}
			if len(onlinePeers) == 1 {
				myElevator.SetOnline(false)
			}

		//RECEIVE FROM NETWORK
		case msg := <-receivedMessages:
			peerID, _ := strconv.Atoi(msg.ID)
			if peerID != id {
				fmt.Println("MSG RECEIVED")
			}

			if sysState == Master && msg.Origin == MO_Slave {
				slaveInfo := ExtractSlaveInformation(msg)
				newElev := Elevator{
					direction:    slaveInfo.direction,
					currentFloor: slaveInfo.currentFloor,
					obs:          slaveInfo.obs,
					priOrder:     elevatorPeers[peerID].priOrder,
					index:        peerID,
				}
				fmt.Println(peerID, "is at", slaveInfo.currentFloor)
				elevatorPeers[peerID] = &newElev
				for _, ord := range slaveInfo.CompletedOrders {
					SetOrder(&MasterOrderPanel, ord, OT_NoOrder, peerID)
				}
				for _, ord := range slaveInfo.NewOrders {
					SetOrder(&MasterOrderPanel, ord, OT_Order, peerID)
				}

			} else if sysState == Slave && msg.Origin == MO_Master {
				masterInfo := ExtractMasterInformation(msg, NUMBER_OF_FLOORS, NUMBER_OF_COLUMNS, NUMBER_OF_ELEVATORS)
				if msg.ID != strconv.Itoa(id) {
					MasterOrderPanel = masterInfo.OrderPanel
					fmt.Println(MasterOrderPanel)
				}

				var compOrdersUpdate []elevio.ButtonEvent
				for _, ord := range newOrders {
					if GetOrder(MasterOrderPanel, ord, peerID) != OT_NoOrder {
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
			//elevatorPeers[elevIndex] = &myElevator
			switch sysState {
			case Master:
				for _, ord := range newOrders {
					SetOrder(&MasterOrderPanel, ord, OT_Order, myElevator.GetIndex())
				}
				newOrders = []elevio.ButtonEvent{}
				for _, ord := range completeOrders {
					SetOrder(&MasterOrderPanel, ord, OT_NoOrder, myElevator.GetIndex())
				}
				completeOrders = []elevio.ButtonEvent{}
				var myElevatorlist []Elevator = []Elevator{myElevator}

				currentPriOrder := myElevator.GetPriOrder()
				myElevatorlist = PrioritizeOrders(MasterOrderPanel, myElevatorlist)
				myElevator = myElevatorlist[0]
				newPriOrder := myElevator.GetPriOrder()

				if currentPriOrder.Floor != -1 && newPriOrder != currentPriOrder {
					SetOrder(&MasterOrderPanel, currentPriOrder, OT_Order, myElevator.GetIndex())
				}
				if newPriOrder.Floor != -1 && newPriOrder != currentPriOrder {
					SetOrder(&MasterOrderPanel, newPriOrder, OT_InProgress, myElevator.GetIndex())
				}

				var available []Elevator
				for _, elev := range elevatorPeers {
					if elev.GetOnline() == false {
						available = append(available, *elev)
					}
				}
				available = PrioritizeOrders(MasterOrderPanel, available)
				for _, elev := range available {
					elevatorPeers[elev.GetIndex()].SetPriOrder(elev.GetPriOrder())
				}
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
				fmt.Println("MASTER MSG SENT")
				msgTx <- NewMasterMessage(strconv.Itoa(id), masterInfo)
			case Slave:
				slaveInfo := SlaveInformation{
					direction:       myElevator.GetDirection(),
					currentFloor:    myElevator.GetCurrentFloor(),
					obs:             myElevator.GetObs(),
					NewOrders:       newOrders,
					CompletedOrders: completeOrders,
				}
				fmt.Println("SLAVE MSG SENT")
				msgTx <- NewSlaveMessage(strconv.Itoa(id), slaveInfo)
				//If it's not online it needs to handle it's own prioritized order same as master
				if myElevator.GetOnline() == false {

					for _, ord := range newOrders {
						SetOrder(&MasterOrderPanel, ord, OT_Order, myElevator.GetIndex())
					}
					newOrders = []elevio.ButtonEvent{}
					for _, ord := range completeOrders {
						SetOrder(&MasterOrderPanel, ord, OT_NoOrder, myElevator.GetIndex())
					}
					completeOrders = []elevio.ButtonEvent{}
					var myElevatorlist []Elevator = []Elevator{myElevator}

					currentPriOrder := myElevator.GetPriOrder()
					myElevatorlist = PrioritizeOrders(MasterOrderPanel, myElevatorlist)
					myElevator = myElevatorlist[0]
					newPriOrder := myElevator.GetPriOrder()

					if currentPriOrder.Floor != -1 && newPriOrder != currentPriOrder {
						SetOrder(&MasterOrderPanel, currentPriOrder, OT_Order, myElevator.GetIndex())
					}
					if newPriOrder.Floor != -1 && newPriOrder != currentPriOrder {
						SetOrder(&MasterOrderPanel, newPriOrder, OT_InProgress, myElevator.GetIndex())
					}
					//fmt.Println("Actual order:", myElevator.GetPriOrder())
					//TESTING PRINTING
					// if newPriOrder != currentPriOrder {
					// 	fmt.Println("Actual order:", myElevator.GetPriOrder())
					// }

					// if MasterOrderPanel != panel {
					// 	fmt.Println("MASTER_ORDER_PANEL: ", MasterOrderPanel)
					// }
				}
			}
		}
		time.Sleep(PERIOD)
	}
}
