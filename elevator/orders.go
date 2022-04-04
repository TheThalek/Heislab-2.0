package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
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
	CT_DoubleDirSwitchCost = 10000
	CT_ObsCost             = 100000
	CT_StayingPut          = 1000000
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
	orderFloor := order.Floor
	orderDirection := 0

	cost += CT_DistanceCost * intAbs(orderFloor-elevFloor)

	if orderFloor == elevFloor {
		cost += CT_MinCost
	}

	if orderFloor == -1 {
		cost += CT_StayingPut
	}

	if elevFloor < orderFloor {
		orderDirection = int(elevio.MD_Up)
	} else if elevFloor > orderFloor {
		orderDirection = int(elevio.MD_Down)
	}

	if orderDirection != int(elevDirection) && elevDirection != 0 {
		cost += CT_DirSwitchCost
	}
	// fmt.Println("ORDER DIR", orderDirection, "ELEV DIR", elevDirection)

	// //buttonDirection := orderDirection
	// if order.Button == elevio.BT_HallUp {
	// 	buttonDirection = int(elevio.MD_Up)
	// } else if order.Button == elevio.BT_HallDown {
	// 	buttonDirection = int(elevio.MD_Down)
	// }

	// if orderDirection != int(elevDirection) {
	// 	cost += 2 * CT_DirSwitchCost
	// 	if buttonDirection != orderDirection {
	// 		cost += CT_DoubleDirSwitchCost
	// 		//cost -= CT_DistanceCost * intAbs(orderFloor-elevFloor)
	// 		// } else {
	// 		// 	cost += CT_DistanceCost * intAbs(orderFloor-elevFloor)
	// 	}
	// } else if buttonDirection != orderDirection {
	// 	cost += CT_DirSwitchCost
	// 	// 	cost -= CT_DistanceCost * intAbs(orderFloor-elevFloor)
	// 	// } else {
	// 	// 	cost += CT_DistanceCost * intAbs(orderFloor-elevFloor)
	// }
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
	//fmt.Println("MASTERORDER PANEL:", MasterOrderPanel)
	for sliceIndex, elevator := range availableElevators {
		elvIndex := elevator.GetIndex()
		oldOrder := elevator.GetPriOrder()
		oldOrderCost := calculateOrderCost(oldOrder, elevator)

		for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
			var btnColumns = []int{0, 1, elvIndex + 2} //Check for the columns: Up, Down, and the given elevator
			for _, btn := range btnColumns {

				if MasterOrderPanel[floor][btn] == OT_Order { // THIS PART MEANS THAT ELEVATORS CAN STEAL CALLS || MasterOrderPanel[floor][btn] == OT_InProgress {

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
								fmt.Println("FOR ELEVATOR", elv, "THE COST IS", cmprCost)
								if cmprCost < orderCost {
									lowestCostAllElevators = cmprCost
									break
								}
							}
						}

						//if orderCost == lowestCostAllElevators {
						if orderCost == lowestCostAllElevators && order != oldOrder {

							elevator.SetPriOrder(order)

							// fmt.Println("NewORDER:", order, "FOR ELEVATOR", elevator)

							// fmt.Println("I'm old:", oldOrder, "I'm new:", order)

							// fmt.Println("ELEV DIR:", elevator.GetDirection())

							SetOrder(MasterOrderPanel, order, OT_InProgress, elevator.GetIndex())

							availableElevators[sliceIndex] = elevator
						}
						//fmt.Println("MASTER PANEL", MasterOrderPanel)
						fmt.Println("OLD ORDER COST:", oldOrderCost, "NEW ORDER COST:", orderCost)
					}
				}
			}
		}
	}
	for fl := 0; fl < NUMBER_OF_FLOORS; fl++ {
		for col := 0; col < NUMBER_OF_COLUMNS; col++ {
			button := col
			if col > 1 {
				button = 2
			}
			order := elevio.ButtonEvent{
				Floor:  fl,
				Button: elevio.ButtonType(button),
			}
			for _, elev := range availableElevators {
				priOrder := elev.GetPriOrder()
				if GetOrder(*MasterOrderPanel, order, elev.GetIndex()) == OT_InProgress && order != priOrder && priOrder.Floor != -1 {
					//fmt.Println("UNCLAIMING ORDER", order)
					SetOrder(MasterOrderPanel, order, OT_Order, elev.GetIndex())
				}
			}
		}
	}
	if len(availableElevators) > 1 {
		for i := 0; i < len(availableElevators); i++ {
			for j := 0; j < len(availableElevators); j++ {
				if i != j && availableElevators[i].GetPriOrder() == availableElevators[j].GetPriOrder() {
					invalidOrder := elevio.ButtonEvent{Floor: -1}
					availableElevators[j].SetPriOrder(invalidOrder)
				}
			}
		}
	}

	//or return list of priority orders -->available elevators
	//fmt.Println("AVAILABLE:", availableElevators)
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
	//fmt.Println("SETTING ORDER", order, "to", OrderType)//
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
}

func CheckOrderTimeout(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, myElevatorList [NUMBER_OF_ELEVATORS]*Elevator) {
	var inProgressTimers []*time.Time
	var inProgressOrders []elevio.ButtonEvent
	timeout := 2 * NUMBER_OF_ELEVATORS * time.Second
	for {
		for i := 0; i < NUMBER_OF_FLOORS; i++ {
			for j := 0; j < NUMBER_OF_COLUMNS; j++ {
				btn := j
				if btn > 2 {
					btn = 2
				}
				order := elevio.ButtonEvent{
					Floor:  i,
					Button: elevio.ButtonType(btn),
				}
				if GetOrder(*MasterOrderPanel, order, 0) == OT_InProgress && !orderIsInSlice(order, inProgressOrders) {
					inProgressOrders = append(inProgressOrders, order)
					currentTime := time.Now()
					inProgressTimers = append(inProgressTimers, &currentTime)
				}
			}
		}

		var orderTimersUpdate []*time.Time
		var ordersUpdate []elevio.ButtonEvent
		for index, t := range inProgressTimers {
			if time.Since(*t) > timeout {
				//fmt.Println("ORDER TIMEOUT!", inProgressOrders[index])
				SetOrder(MasterOrderPanel, inProgressOrders[index], OT_Order, 0)
				for i := 0; i < len(myElevatorList); i++ {
					if myElevatorList[i].GetPriOrder() == inProgressOrders[index] {
						myElevatorList[i].SetOnline(false)
					}
				}

			} else if GetOrder(*MasterOrderPanel, inProgressOrders[index], 0) != OT_NoOrder {
				orderTimersUpdate = append(orderTimersUpdate, inProgressTimers[index])
				ordersUpdate = append(ordersUpdate, inProgressOrders[index])
			}

		}
		inProgressOrders = ordersUpdate
		inProgressTimers = orderTimersUpdate

		time.Sleep(PERIOD)
	}
}

func RestoreOnline(myElevatorList [NUMBER_OF_ELEVATORS]*Elevator) {
	for {
		for _, elev := range len(myElevatorList) {
			elev.SetOnline(true)
			time.Sleep(10 * time.Second)
		}
	}
}

func orderIsInSlice(ord elevio.ButtonEvent, orderSlice []elevio.ButtonEvent) bool {
	for _, o := range orderSlice {
		if o == ord {
			return true
		}
	}
	return false
}
