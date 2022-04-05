package main

import (
	"Driver-go/elevio"
	"time"
)

const (
	//OT = OrderType
	OT_NoOrder    = 0
	OT_Order      = 1
	OT_InProgress = 2
)
const (
	//CT = COST
	CT_MinCost = -10000
	CT_DistanceCost        = 10
	CT_DirSwitchCost       = 100
	CT_DoubleDirSwitchCost = 10000 //CHANGE: NOT USED
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
	for sliceIndex, elevator := range availableElevators {
		elvIndex := elevator.GetIndex()
		oldOrder := elevator.GetPriOrder()
		CostOldOrder := calculateOrderCost(oldOrder, elevator)

		for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
			var btnColumns = []int{0, 1, elvIndex + 2} //Button Columns neccesary to check: Up, Down, and the given elevator.
			for _, btn := range btnColumns {
				if MasterOrderPanel[floor][btn] == OT_Order { 
					var button int = btn 
					if btn > BT_Cab {
						button = BT_Cab
					}
					orderFound := elevio.ButtonEvent{
						Floor:  floor,
						Button: elevio.ButtonType(button),
					}
					var costOrderFound int = calculateOrderCost(orderFound, elevator)
					if costOrderFound < CostOldOrder {
						var minCostAllElevators int = costOrderFound
						//Only compare with other elevators if it's not a cab-call
						if btn != elvIndex+2 { 
							for _, elv := range availableElevators {
								compareCost := calculateOrderCost(orderFound, elv)
								if compareCost < costOrderFound {
									minCostAllElevators = compareCost
									break
								}
							}
						}
						if costOrderFound == minCostAllElevators && orderFound != oldOrder {
							//Order found replaces old order
							elevator.SetPriOrder(orderFound)
							SetOrder(MasterOrderPanel, orderFound, OT_InProgress, elevator.GetIndex())
							availableElevators[sliceIndex] = elevator
						}
					}
				}
			}
		}
	}
	//CHANGE: I don't know what's happening here
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
}

func CheckOrderTimeout(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, myElevatorList [NUMBER_OF_ELEVATORS]*Elevator) {
	var inProgressTimers []*time.Time
	var inProgressOrders []elevio.ButtonEvent
	timeout := 4 * time.Second
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
		time.Sleep(20 * time.Second)
		for _, elev := range myElevatorList {
			elev.SetOnline(true)
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
