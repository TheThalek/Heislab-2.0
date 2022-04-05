package main

import (
	"Driver-go/elevio"
	"time"
)

const (
	//OT = Order Type
	OT_NoOrder    = 0
	OT_Order      = 1
	OT_InProgress = 2
)
const (
	//CT = Cost Type
	CT_MinCost       = -10000
	CT_DistanceCost  = 10
	CT_DirSwitchCost = 100
	CT_ObsCost       = 100000
	CT_StayingPut    = 1000000
)

func intAbs(x int) int {
	if x < 0 {
		x = -x
	}
	return x
}

func calculateOrderCost(order elevio.ButtonEvent, elevator Elevator) int {
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
	elevObstruct := elevator.GetObs()
	if elevObstruct {
		cost += CT_ObsCost
	}

	return cost
}

func PrioritizeOrders(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, availableElevators []Elevator) []Elevator {
	for sliceIndex, elev := range availableElevators {
		elevIndex := elev.GetIndex()
		orderOld := elev.GetPriOrder()
		costOrderOld := calculateOrderCost(orderOld, elev)

		for fl := 0; fl < NUMBER_OF_FLOORS; fl++ {
			var btnColumns = []int{0, 1, elevIndex + 2} //Button Columns neccesary to check: Up, Down, and the given elevator.
			for _, btn := range btnColumns {
				if MasterOrderPanel[fl][btn] == OT_Order {
					var button int = btn
					if btn > elevio.BT_Cab {
						button = elevio.BT_Cab
					}
					orderFound := elevio.ButtonEvent{
						Floor:  fl,
						Button: elevio.ButtonType(button),
					}
					var costOrderFound int = calculateOrderCost(orderFound, elev)
					if costOrderFound < costOrderOld {
						var minCostAllElevators int = costOrderFound
						//Only compare with other elevators if it's not a cab-call
						if btn != elevIndex+2 {
							for _, compareElev := range availableElevators {
								compareCost := calculateOrderCost(orderFound, compareElev)
								if compareCost < costOrderFound {
									minCostAllElevators = compareCost
									break
								}
							}
						}
						if costOrderFound == minCostAllElevators && orderFound != orderOld {
							elev.SetPriOrder(orderFound)
							SetOrder(MasterOrderPanel, orderFound, OT_InProgress, elev.GetIndex())
							availableElevators[sliceIndex] = elev
						}
					}
				}
			}
		}
	}
	//Remove duplicate In-progress orders from MasterOrderPanel
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
					SetOrder(MasterOrderPanel, order, OT_Order, elev.GetIndex())
				}
			}
		}
	}
	//Remove duplicate priority orders from elevators
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

	return availableElevators
}

func GetOrder(MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, order elevio.ButtonEvent, index int) int {
	var fl int = order.Floor
	var btn int
	if order.Button == elevio.BT_HallUp {
		btn = 0
	} else if order.Button == elevio.BT_HallDown {
		btn = 1
	} else {
		btn = 2 + index
	}
	return MasterOrderPanel[fl][btn]
}

func SetOrder(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, order elevio.ButtonEvent, OrderType int, elevIndex int) {
	var fl int = order.Floor
	var btn int
	if order.Button == elevio.BT_HallUp {
		btn = 0
	} else if order.Button == elevio.BT_HallDown {
		btn = 1
	} else {
		btn = 2 + elevIndex
	}
	MasterOrderPanel[fl][btn] = OrderType
}

func CheckOrderTimeout(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, myElevatorList [NUMBER_OF_ELEVATORS]*Elevator) {
	var inProgressTimers []*time.Time
	var inProgressOrders []elevio.ButtonEvent
	orderTimeLimit := 4 * time.Second
	for {
		for fl := 0; fl < NUMBER_OF_FLOORS; fl++ {
			for btn := 0; btn < NUMBER_OF_COLUMNS; btn++ {
				button := btn
				if button > 2 {
					button = 2
				}
				order := elevio.ButtonEvent{
					Floor:  fl,
					Button: elevio.ButtonType(button),
				}
				if GetOrder(*MasterOrderPanel, order, 0) == OT_InProgress && !isOrderInSlice(order, inProgressOrders) {
					inProgressOrders = append(inProgressOrders, order)
					currentTime := time.Now()
					inProgressTimers = append(inProgressTimers, &currentTime)
				}
			}
		}
		//Remove timedout orders
		var inProgressTimersUpdate []*time.Time
		var inProgressOrdersUpdate []elevio.ButtonEvent
		for index, t := range inProgressTimers {
			if time.Since(*t) > orderTimeLimit {
				SetOrder(MasterOrderPanel, inProgressOrders[index], OT_Order, 0)
				for i := 0; i < len(myElevatorList); i++ {
					if myElevatorList[i].GetPriOrder() == inProgressOrders[index] {
						myElevatorList[i].SetAvilable(false)
					}
				}

			} else if GetOrder(*MasterOrderPanel, inProgressOrders[index], 0) != OT_NoOrder {
				inProgressTimersUpdate = append(inProgressTimersUpdate, inProgressTimers[index])
				inProgressOrdersUpdate = append(inProgressOrdersUpdate, inProgressOrders[index])
			}

		}
		inProgressTimers = inProgressTimersUpdate
		inProgressOrders = inProgressOrdersUpdate

		time.Sleep(PERIOD)
	}
}

func RestoreAvailability(myElevatorList [NUMBER_OF_ELEVATORS]*Elevator) {
	for {
		time.Sleep(20 * time.Second)
		for _, elev := range myElevatorList {
			elev.SetAvilable(true)
		}
	}
}

func isOrderInSlice(ord elevio.ButtonEvent, orderSlice []elevio.ButtonEvent) bool {
	for _, o := range orderSlice {
		if o == ord {
			return true
		}
	}
	return false
}
