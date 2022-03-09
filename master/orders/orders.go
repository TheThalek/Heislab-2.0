package orders

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
	"time"
)

const ConstNumFloors int = 4

const (
	//OT = OrderType
	OT_NoOrder = 0
	OT_Order   = 1
)
const (
	//CT = CostType
	CT_DistanceCost        = 150
	CT_DirSwitchCost       = 100
	CT_DoubleDirSwitchCost = 1000
)

func UpdateOrders(orderPanel *[ConstNumFloors][3]int, receiver <-chan elevio.ButtonEvent) {
	//Updates orderPanel matrix when receiver channel gets button calls
	for {
		order := <-receiver
		SetOrder(orderPanel, order.Floor, int(order.Button), OT_Order)
	}
}

func GetOrder(orderPanel *[ConstNumFloors][3]int, floor int, button int) int {
	return orderPanel[floor][button]
}

func SetOrder(orderPanel *[ConstNumFloors][3]int, floor int, button int, orderType int) {
	lampValue := (orderType != OT_NoOrder)
	orderPanel[floor][button] = orderType
	elevio.SetButtonLamp(elevio.ButtonType(button), floor, lampValue)
}

func calculateOrderCost(order elevio.ButtonEvent, elevFloor int, elevDirection elevio.MotorDirection) int {
	// Based on costed scenarios: on the order floor,above or below floor, type of requirede turns - calculate the cost of the given order
	var cost int = 0
	if order.Floor == elevFloor && ((order.Button == elevio.BT_HallUp && elevDirection == elevio.MD_Up) || (order.Button == elevio.BT_HallDown && elevDirection == elevio.MD_Down) || order.Button == elevio.BT_Cab) {
		return cost
	}
	orderFloor := order.Floor
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
			cost -= CT_DistanceCost * intAbs(orderFloor-elevFloor-1)
		} else {
			cost += CT_DistanceCost * intAbs(orderFloor-elevFloor-1)
		}
	} else if newDirection != orderDirection {
		cost += 0.8 * CT_DirSwitchCost
		cost -= CT_DistanceCost * intAbs(orderFloor-elevFloor-1)
	} else {
		cost += CT_DistanceCost * intAbs(orderFloor-elevFloor-1)
	}

	return cost
}

func PriorityOrder(orderPanel *[ConstNumFloors][3]int, elevFloor int, elevDirection elevio.MotorDirection) elevio.ButtonEvent {
	//Calculate for given elevator which order it should take using calculateOrderCost for each current order.
	//fmt.Printf("Y00")
	var priorityOrder elevio.ButtonEvent = elevio.ButtonEvent{
		Floor:  -1,
		Button: -1,
	}
	var minCost int = 100000 //change to infinity <3
	for floor := 0; floor < len(orderPanel); floor++ {
		for btn := 0; btn < len(orderPanel[0]); btn++ {
			if orderPanel[floor][btn] != OT_NoOrder {
				order := elevio.ButtonEvent{
					Floor:  floor,
					Button: elevio.ButtonType(btn),
				}
				//fmt.Println("Order: " + fmt.Sprint(order.Floor) + ", " + fmt.Sprint(order.Button) + " Elevator: " + fmt.Sprint(elevFloor) + ", " + fmt.Sprint(elevDirection))
				orderCost := calculateOrderCost(order, elevFloor, elevDirection)
				if orderCost < minCost {
					minCost = orderCost
					priorityOrder = order
				}
			}
		}
	}
	return priorityOrder
}

func PollPriorityOrder(priOrderChan chan elevio.ButtonEvent, orderPanel *[ConstNumFloors][3]int, myElevator *elevator.Elevator) {
	for {
		order := PriorityOrder(orderPanel, myElevator.GetCurrentFloor(), myElevator.GetDirection())
		if order.Floor != -1 {
			priOrderChan <- order
		}
		time.Sleep(time.Millisecond)
	}
}

func intAbs(x int) int {
	if x < 0 {
		x = -x
	}
	return x
}
