package main

import (
	"Driver-go/elevio"
	//"time"
)

const (
	//OT = OrderType
	OT_NoOrder    = 0
	OT_Order      = 1
	OT_InProgress = 2
)
const (
	CT_MinCost = -10000

	CT_DistanceCost        = 10
	CT_DirSwitchCost       = 100
	CT_DoubleDirSwitchCost = 1000
	CT_ObsCost             = 10000
	CT_StayingPut          = 100000
)

func intAbs(x int) int {
	if x < 0 {
		x = -x
	}
	return x
}

func calculateOrderCost(order elevio.ButtonEvent, elevator Elevator) int {
	// Based on costed scenarios: on the order floor,above or below floor, type of requirede turns - calculate the cost of the given order
	//Add cost of obstruction
	elevFloor := elevator.GetCurrentFloor()
	elevDirection := elevator.GetDirection()
	var cost int = 0
	if order.Floor == elevFloor && ((order.Button == elevio.BT_HallUp && elevDirection == elevio.MD_Up) || (order.Button == elevio.BT_HallDown && elevDirection == elevio.MD_Down) || order.Button == elevio.BT_Cab) {
		return CT_MinCost
	}
	orderFloor := order.Floor
	if orderFloor == -1 {
		return CT_StayingPut
	}

	orderDirection := 0
	if elevFloor < orderFloor {
		orderDirection = int(elevio.MD_Up)
	} else if elevFloor > orderFloor {
		orderDirection = int(elevio.MD_Down)
	}
	newDirection := orderDirection
	if order.Button == elevio.BT_HallUp {
		newDirection = int(elevio.MD_Up)
	} else if order.Button == elevio.BT_HallDown {
		newDirection = int(elevio.MD_Down)
	}

	if orderDirection != int(elevDirection) {
		cost += CT_DirSwitchCost
		if newDirection != orderDirection {
			cost += CT_DoubleDirSwitchCost
			cost -= CT_DistanceCost * intAbs(orderFloor-elevFloor)
		} else {
			cost += CT_DistanceCost * intAbs(orderFloor-elevFloor)
		}
	} else if newDirection != orderDirection {
		cost += 0.8 * CT_DirSwitchCost
		cost -= CT_DistanceCost * intAbs(orderFloor-elevFloor)
	} else {
		cost += CT_DistanceCost * intAbs(orderFloor-elevFloor)
	}
	elevObstruct := elevator.GetObs()
	if elevObstruct {
		cost += CT_ObsCost
	}

	return cost
}

func PrioritizeOrders(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, availableElevators []Elevator) []Elevator {
	//decide which elevator is the best to do an order
	//Need direction for each elevator
	//for each elevator calculate the best order it should take
	//Hvilken ordre er best for hver heis, og hvilken heis er best for ordren.
	//Avaialble elevators assumed sorted. so that elevator 1 comes first in the range, and
	for sliceIndex, elevator := range availableElevators {
		elvIndex := elevator.GetIndex()
		oldOrderCost := calculateOrderCost(elevator.GetPriOrder(), elevator)
		oldOrder := elevator.GetPriOrder()
		for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
			var btnColumns = []int{0, 1, elvIndex + 2} //Check for the columns: Up, Down, and the given elevator
			for _, btn := range btnColumns {

				if MasterOrderPanel[floor][btn] == OT_Order || MasterOrderPanel[floor][btn] == OT_InProgress+elvIndex {

					var button int = btn
					if btn > 1 {
						button = 2
					}
					order := elevio.ButtonEvent{
						Floor:  floor,
						Button: elevio.ButtonType(button),
					}
					var orderCost int = calculateOrderCost(order, elevator)
					if orderCost < oldOrderCost {
						var lowestCostAllElevators int = orderCost
						if btn != elvIndex+2 { //if the btn pushed is not a cab-call, compare with the other elevators
							for _, elv := range availableElevators {
								cmprCost := calculateOrderCost(order, elv)
								if cmprCost < orderCost {
									lowestCostAllElevators = cmprCost
									break
								}
							}
						}

						//if orderCost == lowestCostAllElevators {
						if orderCost == lowestCostAllElevators && oldOrder != order {

							elevator.SetPriOrder(order)

							//fmt.Println("NewORDER:", order)

							//fmt.Println(order)
							//fmt.Println("OLD_ORDER:")
							//fmt.Println(oldOrder)
							//fmt.Println("I'm old:", oldOrder, "I'm new:", order)
							if oldOrder.Floor != -1 {
								SetOrder(MasterOrderPanel, oldOrder, OT_Order, elevator.GetIndex())

								SetOrder(MasterOrderPanel, order, OT_InProgress+elvIndex, elevator.GetIndex())
							} else {
								SetOrder(MasterOrderPanel, order, OT_InProgress+elvIndex, elevator.GetIndex())
								//oldOrder

							}

							availableElevators[sliceIndex] = elevator
						}
						//fmt.Println("NEW ORDER COST:", orderCost)
					}
				}
			}
		}
	}
	//or return list of priority orders -->available elevators
	return availableElevators
}

func GetOrder(MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, order elevio.ButtonEvent, index int) int {
	var fl int = order.Floor
	var bt int
	if order.Button == elevio.BT_HallUp {
		bt = 0
	} else if order.Button == elevio.BT_HallDown {
		bt = 1
	} else {
		bt = 2 + index
	}
	return MasterOrderPanel[fl][bt]
}

func SetOrder(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, order elevio.ButtonEvent, OrderType int, index int) {
	var fl int = order.Floor
	var bt int
	if order.Button == elevio.BT_HallUp {
		bt = 0
	} else if order.Button == elevio.BT_HallDown {
		bt = 1
	} else {
		bt = 2 + index
	}
	MasterOrderPanel[fl][bt] = OrderType
	//fmt.Println("SETTING ORDER:", order)
}
