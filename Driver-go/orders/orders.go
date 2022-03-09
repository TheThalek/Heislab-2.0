package orders

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
	"time"
)

const ConstNumFloors int = 4
const ConstNumElevators int = 3

const (
	//OT = OrderType
	OT_NoOrder    = 0
	OT_Order      = 1
	OT_InProgress = 2
)
const (
	//CT = CostType
	CT_DistanceCost        = 10
	CT_DirSwitchCost       = 100
	CT_DoubleDirSwitchCost = 1000
	CT_ObsCost = 10000
)


//ToDo: Calculate cost for each individual elevator
//Orderpanel[] (constNumFloors-1) x (2 + constNumElevators -1)

func UpdateMasterOrderPanel (){
	//Todo get orders from the slave elevators and update the matrix

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

	return cost
}



func PrioritizeOrders(MasterOrderPanel * [ConstNumFloors][ConstNumElevators+2], availableElevators [Elevator] type ){
	//decide which elevator is the best to do an order
	//Need direction for each elevator 
	//for each elevator calculate the best order it should take
	//Hvilken ordre er best for hver heis, og hvilken heis er best for ordren. 
	test = 1;
	for floor:=0; floor <len(ConstNumFloors); floor ++{
		for btn:= 0; btn < len(ConstNumElevators+2):btn ++{
			if MasterOrderPanel[floor][btn] = OT_Order{
				order := elevio.ButtonEvent{
					Floor: floor,
					Button: elevio.ButtonType(btn),
				}
				var orderminimumcost int = 100000
				betterElevator = nil
				for elevator in availableElevators{
					orderCostelevator := calculateOrderCost(order,elevator.currentFloor, elevator.direction)
					if orderCostelevator < orderminimumcost{
						//check if it can overwrite current order
						orderminimumcost = orderCostelevator
						elevatorCurrentOrderCost :=calculateOrderCost(elevator.priOrder, elevator.currentFloor, elevator.direction)
						if orderCostelevator < elevatorCurrentOrderCost{
							betterElevator  = elevator
						}
					}
				}
				//update order if there is a better option
				if betterElevator{
					//overwrite current order, set the value of the order to OT_order from OT_progress or smth
					MasterOrderPanel[]
				}

			}
		}
	}

	//return elv_1, elv_2, elv_3
}



///GAMMEL:
func PriorityOrder(orderPanel *[ConstNumFloors][3]int, elevFloor int, elevDirection elevio.MotorDirection) elevio.ButtonEvent {
	//Calculate for given elevator which order it should take using calculateOrderCost for each current order.
	var priorityOrder elevio.ButtonEvent = elevio.ButtonEvent{
		Floor:  -1,
		Button: -1,
	}
	var minCost int = 100000 //change to infinity
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

func PollPriorityOrder(priOrderChan <-chan elevio.ButtonEvent, orderPanel *[ConstNumFloors][3]int, myElevator *elevator.Elevator) {
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
