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

func calculateOrderCost(order elevio.ButtonEvent, elevator elevator.Elevator) int {
	// Based on costed scenarios: on the order floor,above or below floor, type of requirede turns - calculate the cost of the given order
	//Add cost of obstruction
	elevFloor = elevator.currentFloor
	elevDirection = elevator.direction
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
	if elevator.obs{
		cost += CT_ObsCost
	}
	
	return cost
}



func PrioritizeOrders(masterOrderPanel * [ConstNumFloors][ConstNumElevators+2]int, availableElevators [Elevator] type ){
	//decide which elevator is the best to do an order
	//Need direction for each elevator 
	//for each elevator calculate the best order it should take
	//Hvilken ordre er best for hver heis, og hvilken heis er best for ordren. 
	//Avaialble elevators assumed sorted. so that elevator 1 comes first in the range, and
	for elvIndex, elevator := range availableElevators{
		oldOrderCost:= calculateOrderCost(elevator.priOrder, elevator)
		oldOrder:= elevator.priOrder
		for floor:= 0; floor < len(ConstNumFloors); floor ++{
			var btnColumns= []int{0,1,elvIndex+2} //Check for the columns: Up, Down, and the given elevator
			for _, btn := range btnColumns{
				if MasterOrderPanel[floor][btn] == OT_Order{
					order := elevio.ButtonEvent{
						Floor: floor,
						Button: elevio.ButtonType(btn),
					}
					orderCost = calculateOrderCost(order,elevator)
					if orderCost < oldOrderCost{
						var lowestCostAllElevators int = orderCost
						if btn != elvIndex+2{ //if the btn pushed is not a cab-call, compare with the other elevators
							for _, elv:= range availableElevators{
								cmprCost := calculateOrderCost(elv.order, elv)
								if cmprCost < orderCost{
									lowestCostAllElevators = cmprCost
									break
								}
							}
						}
						if orderCost == lowestCostAllElevators{
							elevator.priOrder = order
							masterOrderPanel[oldOrder.Floor][oldOrder.Button] = OT_Order
							masterOrderPanel[order.Floor][order.Button] = OT_InProgress
						} 

					}
				}
			} 
			
		} 

	}
	//and return list of priority orders -->available elevators

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
