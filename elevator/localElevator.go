package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

func ThaleSinMain() {
	LocalInit()
	var masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var myElevator Elevator
	startPri := elevio.ButtonEvent{
		Floor:  -1,
		Button: elevio.ButtonType(0),
	}
	myElevator.SetPriOrder(startPri)

	ElevIndexChan := make(chan int)
	takenOrdersChan := make(chan []elevio.ButtonEvent)
	newOrdersChan := make(chan elevio.ButtonEvent)

	go LocalControl(&myElevator, masterOrderPanel, takenOrdersChan, newOrdersChan, ElevIndexChan)
	for {
		select {
		case t := <-takenOrdersChan:
			fmt.Println(t)
		case t := <-newOrdersChan:
			fmt.Println(t)
		case t := <-ElevIndexChan:
			fmt.Println(t)
		default:
		}
	}
}

func LocalInit() {
	elevio.Init("localhost:15657", NUMBER_OF_FLOORS)
	//elevio.SetMotorDirection(elevio.MD_Down)
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func setLights(masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, elevIndex chan int) {
	for {
		elevatorIndx := <-elevIndex
		for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
			var btnColumns = []int{0, 1, elevatorIndx + 2}
			for _, btn := range btnColumns {
				if masterOrderPanel[floor][btn] == OT_NoOrder {
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
				} else if masterOrderPanel[floor][btn] == OT_Order {
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
				}
			}
		}
	}
}

func pollPriOrder(priOrder chan elevio.ButtonEvent, myElevator *Elevator) {
	var oldOrder elevio.ButtonEvent = myElevator.GetPriOrder()
	for {
		newOrder := myElevator.GetPriOrder()
		if newOrder != oldOrder {
			oldOrder = newOrder
			priOrder <- newOrder
		}
	}
}

func test(myElevator *Elevator) {
	fmt.Println("before 5 sec 1 ")
	time.Sleep(5 * time.Second)
	fmt.Println("after 5 sec 1")
	testPri1 := elevio.ButtonEvent{
		Floor:  1,
		Button: elevio.ButtonType(0),
	}
	myElevator.SetPriOrder(testPri1)

	fmt.Println("before 5 sec 2")
	time.Sleep(15 * time.Second)
	fmt.Println("after 5 sec 2")
	testPri2 := elevio.ButtonEvent{
		Floor:  3,
		Button: elevio.ButtonType(0),
	}
	myElevator.SetPriOrder(testPri2)
}

func LocalControl(myElevator *Elevator, masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, takenOrders chan []elevio.ButtonEvent, newOrders chan elevio.ButtonEvent, elevIndex chan int) {

	elevio.SetMotorDirection(elevio.MD_Down)

	var doorOpen bool = false
	var moving bool = true
	var obs bool = false
	myElevator.SetObs(obs)

	//Kanskje legg inn i myElevator
	var priorityOrder elevio.ButtonEvent
	priorityOrder.Floor = -1
	myElevator.SetPriOrder(priorityOrder)

	go setLights(masterOrderPanel, elevIndex) //TO DO; ha med en "polingsfunksjon" i main

	drv_stop := make(chan bool)
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	priOrderChan := make(chan elevio.ButtonEvent)

	//TEST
	go test(myElevator)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	go pollPriOrder(priOrderChan, myElevator)

	for {
		//OPPDATER ELLER DOBBELTSJEKK AT priorityOrder blir oppdatert
		select {
		case currentOrder := <-priOrderChan:

			myElevator.SetPriOrder(currentOrder)
			if myElevator.GetPriOrder().Floor != myElevator.GetCurrentFloor() && myElevator.GetPriOrder().Floor != -1 {
				//stop moving
				moving = true
				//fmt.Println("floor >> moving")
			}
			if !doorOpen && myElevator.GetPriOrder().Floor != -1 {
				//drive to the order
				if myElevator.GetCurrentFloor() != myElevator.GetPriOrder().Floor {
					myElevator.DriveTo(myElevator.GetPriOrder())
				}
				if !moving && myElevator.GetPriOrder().Floor == myElevator.GetCurrentFloor() {
					if myElevator.GetPriOrder().Button == elevio.BT_HallUp {
						myElevator.SetDirection(elevio.MD_Up)
					} else if myElevator.GetPriOrder().Button == elevio.BT_HallDown {
						myElevator.SetDirection(elevio.MD_Down)
					}
					// create button event corresponding to current elev state
					event := elevio.ButtonEvent{
						Floor:  myElevator.GetCurrentFloor(),
						Button: elevio.BT_Cab,
					}
					if myElevator.GetDirection() == elevio.MD_Up {
						event.Button = elevio.BT_HallUp
					} else if myElevator.GetDirection() == elevio.MD_Down {
						event.Button = elevio.BT_HallDown
					}
					//open doors
					doorOpen = true
					elevio.SetDoorOpenLamp(doorOpen)
					//fmt.Println("pri >> door open")
					//timer
					time.Sleep(3 * time.Second)
					//clear the orders
					var newFloor = myElevator.GetCurrentFloor()
					var completedOrders []elevio.ButtonEvent
					if masterOrderPanel[newFloor][<-elevIndex+2] != OT_NoOrder {
						cabOrder := elevio.ButtonEvent{
							Floor:  newFloor,
							Button: elevio.ButtonType(2),
						}
						completedOrders = append(completedOrders, cabOrder)
					}
					if (masterOrderPanel[newFloor][0] != OT_NoOrder) && (myElevator.GetDirection() == elevio.MD_Up) {
						dirOrder := elevio.ButtonEvent{
							Floor:  newFloor,
							Button: elevio.ButtonType(0),
						}
						completedOrders = append(completedOrders, dirOrder)
					} else if (masterOrderPanel[newFloor][1] != OT_NoOrder) && (myElevator.GetDirection() == elevio.MD_Down) {
						dirOrder := elevio.ButtonEvent{
							Floor:  newFloor,
							Button: elevio.ButtonType(1),
						}
						completedOrders = append(completedOrders, dirOrder)
					}
					takenOrders <- completedOrders

					//set priority to an invalid order
					var priorityOrder elevio.ButtonEvent
					priorityOrder.Floor = -1
					myElevator.SetPriOrder(priorityOrder)
					//open door
					if !obs {
						doorOpen = false
						elevio.SetDoorOpenLamp(doorOpen)
					}
					//fmt.Println("pri >> door closed")
				}
			}

		case newBtnEvent := <-drv_buttons:
			fmt.Println("drv_buttons")
			newOrders <- newBtnEvent
			fmt.Println("after drv_buttons")

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
			myElevator.SetObs(ObstrEvent)
			if ObstrEvent && !moving {
				doorOpen = true
				elevio.SetDoorOpenLamp(doorOpen)
			} else {
				doorOpen = false
			}

		case stopEvent := <-drv_stop:
			fmt.Println(stopEvent)
		}
	}
}
