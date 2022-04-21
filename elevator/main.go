package main

import (
	"Driver-go/elevio"
	"strconv"
)

type SystemState int

const (
	Slave  = 0
	Master = 1
)

func main() {
	//Initiation
	var myElevator Elevator = NewElevator()
	var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int
	var sysState SystemState = Slave

	var completeOrders []elevio.ButtonEvent
	var newOrders []elevio.ButtonEvent
	completeOrderChan := make(chan []elevio.ButtonEvent)
	newOrderChan := make(chan elevio.ButtonEvent)

	LocalElevatorInit()
	go LocalElevatorControl(&myElevator, &MasterOrderPanel, completeOrderChan, newOrderChan)
	myID := NetworkConnect()

	//ID must be a positive integer in the range [0, NUMBER_OF_ELEVATORS)
	elevIndex := myID
	myElevator.SetIndex(elevIndex)

	var elevatorPeers [NUMBER_OF_ELEVATORS]*Elevator
	for i := 0; i < len(elevatorPeers); i++ {
		nElev := NewElevator()
		nElev.SetIndex(i)
		elevatorPeers[i] = &nElev
	}

	elevatorPeers[elevIndex] = &myElevator
	onlinePeers := []int{}

	msgTx := make(chan NetworkMessage)
	receivedMessages := make(chan NetworkMessage, NUMBER_OF_ELEVATORS)
	roleChan := make(chan string)
	peerChan := make(chan []int)

	go RunNetworkInterface(myID, msgTx, receivedMessages, roleChan, peerChan)

	go CheckOrderTimeout(&MasterOrderPanel, elevatorPeers)

	//go RestoreAvailability(elevatorPeers)

	for {
		select {
		case cOrds := <-completeOrderChan:
			completeOrders = append(completeOrders, cOrds...)
		case nOrds := <-newOrderChan:
			newOrders = append(newOrders, nOrds)
		case role := <-roleChan:
			if role == string(MO_Master) {
				sysState = Master
			} else if role == string(MO_Slave) {
				sysState = Slave
			}

		case onlinePeers = <-peerChan:
			for i := 0; i < NUMBER_OF_ELEVATORS; i++ {
				if isIntInSlice(i, onlinePeers) {
					elevatorPeers[i].SetAvilable(true)
				} else {
					elevatorPeers[i].SetAvilable(false)
				}
			}
			if len(onlinePeers) == 1 {
				myElevator.SetAvilable(false)
			}

		//RECEIVE FROM NETWORK
		case msg := <-receivedMessages:
			peerID, _ := strconv.Atoi(msg.ID)
			if peerID != myID {

				if sysState == Master && msg.Origin == MO_Slave {
					slaveInfo := ExtractSlaveInformation(msg)
					newElev := Elevator{
						direction:    slaveInfo.direction,
						currentFloor: slaveInfo.currentFloor,
						obs:          slaveInfo.obs,
						priOrder:     elevatorPeers[peerID].priOrder,
						index:        peerID,
						available:    true,
					}
					elevatorPeers[peerID] = &newElev
					for _, ord := range slaveInfo.NewOrders {
						SetOrder(&MasterOrderPanel, ord, OT_Order, peerID)
					}
					for _, ord := range slaveInfo.CompletedOrders {
						SetOrder(&MasterOrderPanel, ord, OT_NoOrder, peerID)
						invalidOrder := elevio.ButtonEvent{Floor: -1}
						elevatorPeers[peerID].SetPriOrder(invalidOrder)
					}
				} else if sysState == Slave && msg.Origin == MO_Master {
					masterInfo := ExtractMasterInformation(msg, NUMBER_OF_FLOORS, NUMBER_OF_COLUMNS, NUMBER_OF_ELEVATORS)
					MasterOrderPanel = masterInfo.OrderPanel
					var compOrdersUpdate []elevio.ButtonEvent
					for _, ord := range completeOrders {
						if GetOrder(MasterOrderPanel, ord, myID) != OT_NoOrder {
							compOrdersUpdate = append(compOrdersUpdate, ord)
						}
					}
					completeOrders = compOrdersUpdate
					var newOrdersUpdate []elevio.ButtonEvent
					for _, ord := range newOrders {
						if GetOrder(MasterOrderPanel, ord, myID) == OT_NoOrder {
							newOrdersUpdate = append(newOrdersUpdate, ord)
						}
					}
					newOrders = newOrdersUpdate
					myElevator.SetPriOrder(masterInfo.Priorities[myID].order)
				}
			}
		//SEND TO NETWORK
		default:
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

				var available []*Elevator
				for _, elev := range elevatorPeers {
					if elev.GetAvailable() == true {
						available = append(available, elev)
					}
				}
				available = PrioritizeOrders(&MasterOrderPanel, available)

				for _, elev := range available {
					priorityOrder := elev.GetPriOrder()
					index := elev.GetIndex()
					elevatorPeers[index].SetPriOrder(priorityOrder)
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

				msgTx <- NewMasterMessage(strconv.Itoa(myID), masterInfo)
			case Slave:
				slaveInfo := SlaveInformation{
					direction:       myElevator.GetDirection(),
					currentFloor:    myElevator.GetCurrentFloor(),
					obs:             myElevator.GetObs(),
					NewOrders:       newOrders,
					CompletedOrders: completeOrders,
				}
				msgTx <- NewSlaveMessage(strconv.Itoa(myID), slaveInfo)

				//If it's not online it needs to handle it's own prioritized order same as master
				if len(onlinePeers) <= 1 {

					for _, ord := range newOrders {
						SetOrder(&MasterOrderPanel, ord, OT_Order, myElevator.GetIndex())
					}
					newOrders = []elevio.ButtonEvent{}
					for _, ord := range completeOrders {
						SetOrder(&MasterOrderPanel, ord, OT_NoOrder, myElevator.GetIndex())
						invalidOrder := elevio.ButtonEvent{Floor: -1}
						elevatorPeers[myID].SetPriOrder(invalidOrder)
					}
					completeOrders = []elevio.ButtonEvent{}

					var myElevatorlist []*Elevator = []*Elevator{&myElevator}
					myElevatorlist = PrioritizeOrders(&MasterOrderPanel, myElevatorlist)
					myElevator = *myElevatorlist[0]

				}
			}
		}
	}
}
