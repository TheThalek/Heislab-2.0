package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func LocalInit() { //default: 15657 - Set to random then start elevatorserver to elevatorserver --port15054

	elevio.Init("localhost:39417", NUMBER_OF_FLOORS)

	elevio.SetDoorOpenLamp(false)

	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func setLights(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, myElevator *Elevator) {
	for {
		for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
			var btnColumns = []int{0, 1, myElevator.GetIndex() + 2}
			for _, btn := range btnColumns {
				var lightValue bool
				if MasterOrderPanel[floor][btn] == OT_NoOrder {
					lightValue = false
				} else {
					lightValue = true
				}
				var btnType elevio.ButtonType
				if btn == 0 {
					btnType = elevio.BT_HallUp
				} else if btn == 1 {
					btnType = elevio.BT_HallDown
				} else {
					btnType = elevio.BT_Cab
				}
				elevio.SetButtonLamp(btnType, floor, lightValue)
			}
		}
		time.Sleep(PERIOD)
	}
}

func pollPriOrder(priOrder chan elevio.ButtonEvent, myElevator *Elevator) {
	for {
		priOrder <- myElevator.GetPriOrder()
		time.Sleep(PERIOD)
	}
}

func LocalControl(myElevator *Elevator, MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, takenOrders chan []elevio.ButtonEvent, newOrders chan elevio.ButtonEvent) {

	elevio.SetMotorDirection(elevio.MD_Down)

	var doorOpen bool = false
	var moving bool = true
	myElevator.SetObs(false)

	var priorityOrder elevio.ButtonEvent
	priorityOrder.Floor = -1
	myElevator.SetPriOrder(priorityOrder)

	go setLights(MasterOrderPanel, myElevator)

	drv_stop := make(chan bool)
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	priOrderChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go pollPriOrder(priOrderChan, myElevator)

	for {
		select {
		case priorityOrder := <-priOrderChan:
			//fmt.Println("currentOrder", currentOrder)
			var currentFloor int = myElevator.GetCurrentFloor()

			if priorityOrder.Floor != currentFloor && priorityOrder.Floor != -1 {
				//Not on the same floor as the order
				moving = true
			}
			if !doorOpen && priorityOrder.Floor != -1 {
				//Door is closed and you have an order
				if currentFloor != priorityOrder.Floor {
					//Elevator is not on the prioritized floor
					myElevator.DriveTo(priorityOrder)
				}
				if !moving && priorityOrder.Floor == currentFloor {
					//Not moving and elevator is at the correct floor
					//Switch direction to suit the order it's priorityOrder
					if priorityOrder.Button == elevio.BT_HallUp {
						myElevator.SetDirection(elevio.MD_Up)
					} else if priorityOrder.Button == elevio.BT_HallDown {
						myElevator.SetDirection(elevio.MD_Down)
					}

					//open doors

					doorOpen = true
					elevio.SetDoorOpenLamp(doorOpen)
					//timer
					time.Sleep(1500 * time.Millisecond)

					//clear the relevant orders
					var completedOrders []elevio.ButtonEvent
					completedOrders = append(completedOrders, priorityOrder)
					//compltOrder := priorityOrder
					//MAIKEN DU HOLDER PÅ HER!
					// if compltOrder.Button == elevio.BT_Cab{
					// 	compltOrder.Button = elevio.BT_Cab+myElevator.GetIndex()
					// }
					// if MasterOrderPanel[currentFloor][compltOrder.Button] = OT_NoOrder{
					// 	completedOrders = append(completedOrders, priorityOrder)
					// }

					cabOrder := elevio.ButtonEvent{
						Floor:  currentFloor,
						Button: elevio.ButtonType(elevio.BT_Cab),
					}
					upOrder := elevio.ButtonEvent{
						Floor:  currentFloor,
						Button: elevio.ButtonType(elevio.BT_HallUp),
					}
					downOrder := elevio.ButtonEvent{
						Floor:  currentFloor,
						Button: elevio.ButtonType(elevio.BT_HallDown),
					}

					if GetOrder(*MasterOrderPanel, cabOrder, myElevator.GetIndex()) != OT_NoOrder {
						completedOrders = append(completedOrders, cabOrder)
					}
					if GetOrder(*MasterOrderPanel, upOrder, myElevator.GetIndex()) != OT_NoOrder {
						completedOrders = append(completedOrders, upOrder)
					} else if GetOrder(*MasterOrderPanel, downOrder, myElevator.GetIndex()) != OT_NoOrder {
						completedOrders = append(completedOrders, downOrder)
					}
					takenOrders <- completedOrders

					//set priority to an invalid order
					var priorityOrder elevio.ButtonEvent
					priorityOrder.Floor = -1
					myElevator.SetPriOrder(priorityOrder)
					//fmt.Println("Cleared in local elevator", priorityOrder)
					//open door
					if !myElevator.GetObs() {
						doorOpen = false
						elevio.SetDoorOpenLamp(doorOpen)
					}
					//fmt.Println("pri >> door closed")
				}
			}

		case newBtnEvent := <-drv_buttons:
			newOrders <- newBtnEvent

		case newFloor := <-drv_floors:
			//update the floor
			myElevator.SetFloor(newFloor)
			//turn on the floor light
			elevio.SetFloorIndicator(newFloor)

			myElevator.DriveTo(myElevator.GetPriOrder())
			//if this floor has an order
			if myElevator.GetPriOrder().Floor != newFloor && myElevator.GetPriOrder().Floor != -1 {
				//stop moving
				moving = true
				//fmt.Println("floor >> moving")
			} else {
				moving = false
				elevio.SetMotorDirection(elevio.MD_Stop)
				//fmt.Println("floor >> not moving")
			}
			//switch direction if at top or bottom floor
			if myElevator.GetCurrentFloor() == 0 {
				myElevator.SetDirection(elevio.MD_Up)
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else if myElevator.GetCurrentFloor() == NUMBER_OF_FLOORS-1 {
				myElevator.SetDirection(elevio.MD_Down)
				elevio.SetMotorDirection(elevio.MD_Stop)
			}

		case ObstrEvent := <-drv_obstr:
			fmt.Println("OBSTRUCTION:", ObstrEvent)
			myElevator.SetObs(ObstrEvent)
			if ObstrEvent && !moving {
				doorOpen = true
			} else {
				doorOpen = false
				time.Sleep(3 * time.Second)
			}
			elevio.SetDoorOpenLamp(doorOpen)

		case stopEvent := <-drv_stop:
			fmt.Println("STOP:", stopEvent)
		}
		//time.Sleep(PERIOD)
	}
}
