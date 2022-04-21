package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func LocalElevatorInit() {
	//Default localhost: 15657. Directly dependent on connection with the elevatorserver
	//To initiate connection with elevatorserver use: elevatorserver --port 15054
	elevio.Init("localhost:15002", NUMBER_OF_FLOORS)

	elevio.SetDoorOpenLamp(false)
	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func setLights(MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, myElevator *Elevator) {
	for {
		for fl := 0; fl < NUMBER_OF_FLOORS; fl++ {
			var btnColumns = []int{0, 1, myElevator.GetIndex() + 2}
			for _, btn := range btnColumns {
				var lightValue bool
				if MasterOrderPanel[fl][btn] == OT_NoOrder {
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
				elevio.SetButtonLamp(btnType, fl, lightValue)
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

func LocalElevatorControl(myElevator *Elevator, MasterOrderPanel *[NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, completedOrdersChan chan []elevio.ButtonEvent, newOrdersChan chan elevio.ButtonEvent) {

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
	priOrderChan := make(chan elevio.ButtonEvent) //

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go pollPriOrder(priOrderChan, myElevator)

	for {
		select {
		case priorityOrder := <-priOrderChan:
			var currentFloor int = myElevator.GetCurrentFloor()

			if priorityOrder.Floor != currentFloor && priorityOrder.Floor != -1 {
				moving = true
			}
			if !doorOpen && priorityOrder.Floor != -1 {
				if currentFloor != priorityOrder.Floor {
					myElevator.DriveTo(priorityOrder)
				}
				if !moving && priorityOrder.Floor == currentFloor {
					if priorityOrder.Button == elevio.BT_HallUp {
						myElevator.SetDirection(elevio.MD_Up)
					} else if priorityOrder.Button == elevio.BT_HallDown {
						myElevator.SetDirection(elevio.MD_Down)
					}
					doorOpen = true
					elevio.SetDoorOpenLamp(doorOpen)
					time.Sleep(1500 * time.Millisecond)

					var completedOrders []elevio.ButtonEvent
					completedOrders = append(completedOrders, priorityOrder)

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
					completedOrdersChan <- completedOrders

					var priorityOrder elevio.ButtonEvent
					priorityOrder.Floor = -1 //Setting priOrder to invalid floor to avoid re-enter same case
					myElevator.SetPriOrder(priorityOrder)
					if !myElevator.GetObs() {
						doorOpen = false
						elevio.SetDoorOpenLamp(doorOpen)
					}
				}
			}

		case newBtnEvent := <-drv_buttons:
			newOrdersChan <- newBtnEvent

		case newFloor := <-drv_floors:
			myElevator.SetFloor(newFloor)
			elevio.SetFloorIndicator(newFloor)
			
			time.Sleep(50 * time.Millisecond)
			
			myElevator.DriveTo(myElevator.GetPriOrder())
			if myElevator.GetPriOrder().Floor != newFloor && myElevator.GetPriOrder().Floor != -1 {
				moving = true
			} else {
				moving = false
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			if myElevator.GetCurrentFloor() == 0 {
				myElevator.SetDirection(elevio.MD_Up)
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else if myElevator.GetCurrentFloor() == NUMBER_OF_FLOORS-1 {
				myElevator.SetDirection(elevio.MD_Down)
				elevio.SetMotorDirection(elevio.MD_Stop)
			}

		case obstrEvent := <-drv_obstr:
			fmt.Println("OBSTRUCTION:", obstrEvent)
			myElevator.SetObs(obstrEvent)
			if obstrEvent && !moving {
				doorOpen = true
			} else {
				doorOpen = false
				time.Sleep(3 * time.Second)
			}
			elevio.SetDoorOpenLamp(doorOpen)

		case stopEvent := <-drv_stop:
			fmt.Println("STOP:", stopEvent)
		}
	}
}
