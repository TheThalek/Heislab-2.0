package orders

import (
	"Driver-go/elevator"
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
	//CT = CostType
	CT_DistanceCost        = 10
	CT_DirSwitchCost       = 100
	CT_DoubleDirSwitchCost = 1000
	CT_ObsCost = 10000
)


func intAbs(x int) int {
	if x < 0 {
		x = -x
	}
	return x
}

func calculateOrderCost(order elevio.ButtonEvent, elevator elevator.Elevator) int {
	// Based on costed scenarios: on the order floor,above or below floor, type of requirede turns - calculate the cost of the given order
	//Add cost of obstruction
	elevFloor := elevator.GetCurrentFloor()
	elevDirection := elevator.GetDirection()
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
	elevObstruct:= elevator.GetObs()
	if elevObstruct{
		cost += CT_ObsCost
	}
	
	return cost
}



func PrioritizeOrders(MasterOrderPanel * [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int, availableElevators [elevator.NUMBER_OF_ELEVATORS] elevator.Elevator ){
	//decide which elevator is the best to do an order
	//Need direction for each elevator 
	//for each elevator calculate the best order it should take
	//Hvilken ordre er best for hver heis, og hvilken heis er best for ordren. 
	//Avaialble elevators assumed sorted. so that elevator 1 comes first in the range, and
	for elvIndex, elevator := range availableElevators{
		oldOrderCost:= calculateOrderCost(elevator.GetPriOrder(), elevator)
		oldOrder:= elevator.GetPriOrder()
		for floor:= 0; floor < elevator.NUMBER_OF_FLOORS; floor ++{
			var btnColumns= []int{0,1,elvIndex+2} //Check for the columns: Up, Down, and the given elevator
			for _, btn := range btnColumns{
				if MasterOrderPanel[floor][btn] == OT_Order{
					order := elevio.ButtonEvent{
						Floor: floor,
						Button: elevio.ButtonType(btn),
					}
					var orderCost int = calculateOrderCost(order,elevator)
					if orderCost < oldOrderCost{
						var lowestCostAllElevators int = orderCost
						if btn != elvIndex+2{ //if the btn pushed is not a cab-call, compare with the other elevators
							for _, elv:= range availableElevators{
								cmprCost := calculateOrderCost(order, elv)
								if cmprCost < orderCost{
									lowestCostAllElevators = cmprCost
									break
								}
							}
						}
						if orderCost == lowestCostAllElevators{
							elevator.SetPriOrder(order)
							MasterOrderPanel[oldOrder.Floor][oldOrder.Button] = OT_Order
							MasterOrderPanel[order.Floor][order.Button] = OT_InProgress
						} 

					}
				}
			} 
			
		} 

	}
	//or return list of priority orders -->available elevators
	return availableElevators

}



//Make a function that get's the timeout message, and then sets and order back to OT_order

func TimeOutElevatorOrder(MasterOrderPanel * [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int, lostElevator elevator.Elevator){
	lostOrder = lostElevator.GetPriOrder() 
	if lostOrder.Button != BT_Cab{
		MasterOrderPanel[lostOrder.Floor][lostOrder.Button] = OT_Order
	}
}


//Set an order back to complete when it's been done, or if it's d
//
func CompletedOrder(MasterOrderPanel * [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int, completeElevator elevator.Elevator){
	completeOrder = completeElevator.GetPriOrder()
	MasterOrderPanel[completeOrder.Floor][completeOrder.Button] = OT_NoOrder
}

func maikenSinMain(){
	//Test av Orders funksjoner
	var elevator_1 elevator.Elevator
	elevator_1.direction = 1
	elevator_1.currentFloor = 1
	elevator_1.obs = false
	elevator_1.priOrder.Floor = 1
	elevator_1.priOrder.Button = 1

	var elevator_2 elevator.Elevator
	elevator_2.direction = 1
	elevator_2.currentFloor = 1
	elevator_2.obs = false
	elevator_2.priOrder.Floor = 1
	elevator_2.priOrder.Button = 1

	//Testing scenario: 



}


//Fra Single Elevator - Peder:
func PollPriorityOrder(priOrderChan <-chan elevio.ButtonEvent, orderPanel *[elevator.NUMBER_OF_FLOORS][3]int, myElevator *elevator.Elevator) {
	for {
		order := PriorityOrder(orderPanel, myElevator.GetCurrentFloor(), myElevator.GetDirection())
		if order.Floor != -1 {
			priOrderChan <- order
		}
		time.Sleep(time.Millisecond)
	}
}

