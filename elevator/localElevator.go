package main

import (
	"Driver-go/elevio"
	"fmt"
	"time"
)

type SlaveState int

const (
	Idle        SlaveState = 0
	Move                   = 1
	Obstruction            = 2
)

func ThaleSinMain() {
	SlaveFSMinit()
	fmt.Println("Test")
	var masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var localElevator Elevator

	ElevIndexChan := make(chan int)
	takenOrdersChan := make(chan []elevio.ButtonEvent)
	newOrdersChan := make(chan []elevio.ButtonEvent)

	go SlaveFSM(&localElevator, masterOrderPanel, takenOrdersChan, newOrdersChan, ElevIndexChan)

	for {
	}
}

func SlaveFSMinit() {
	elevio.Init("localhost:15657", NUMBER_OF_FLOORS)

	elevio.SetMotorDirection(elevio.MD_Down)

	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

//MAIKEN HOPPER INN FOR DENNE:
func setLights(masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, elevIndex int) {
	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			if (b == 0) || (b == 1) { //If up or down pushed
				elevio.SetButtonLamp(elevio.ButtonType(b), f, (masterOrderPanel[f][b] != OT_NoOrder)) //Will set the lamp on/off if 0/1or2
			} else if b == 2 { //If cab
				elevio.SetButtonLamp(elevio.ButtonType(b), f, (masterOrderPanel[f][elevIndex+2]) != OT_NoOrder) //GetElevatorIndex gives the nr. of column
			}
		}
	}
}

//THALE JOBBER M RESTEN
func pollPriFloor(priOrder chan elevio.ButtonEvent, localElevator Elevator) {
	for {
		priOrder <- localElevator.GetPriOrder()
	}
}

func SlaveFSM(localElevator *Elevator, masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, takenOrders chan []elevio.ButtonEvent, newOrders chan []elevio.ButtonEvent, elevIndex chan int) {

	var state SlaveState = Idle
	var currentDirection = elevio.MD_Down
	localElevator.setDirection(currentDirection)

	//Og kjøre nedover til den når den nederste etasjen sin!

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	priOrderChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go pollPriFloor(<-priOrderChan, localElevator)
	go setLights(masterOrderPanel, <-elevIndex) //TO DO; ha med en "polingsfunksjon" i main

	for {
		if SlaveState == Move {
			//Køyr til etasjen du skal til OG du må endre direction du går i (i localElevator), dersom du endrer denne!
			driveTo(&localElevator)
		}

		select {
		case obstr := <-drv_obstr:
			switch {
			case SlaveState == Move:
				elevio.SetMotorDirection(elevio.MD_Stop)
				localElevator.setObs(true)
				SlaveState = Obstruction
			case SlaveState == Idle:
				localElevator.setObs(true)
				SlaveState = Obstruction
			case SlaveState == Obstruction:
				if localElevator.GetPriOrder() == OT_NoOrder {
					localElevator.setobs(false)
					SlaveState = Idle
				} else {
					localElevator.setobs(false)
					SlaveState = Move
				}
			}

		case newFloor := <-drv_floors:
			localElevator.SetFloor(newFloor)
			SetFloorIndicator(newFloor)

			if newFloor == localElevator.GetPriOrder().Floor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				SetDoorOpenLamp(true)
				time.Sleep(3 * time.Second)
				SetDoorOpenLamp(false)

				var completedOrders []ButtonEvent
				if masterOrderPanel[newFloor][<-elevIndex+2] != OT_NoOrder {
					cabOrder := elevio.ButtonEvent{
						Floor:  newFloor,
						Button: elevio.ButtonType(2),
					}
					completedOrders = append(completedOrders, cabOrder)
				}
				if (masterOrderPanel[newFloor][0] != OT_NoOrder) & (localElevator.GetDirection() == MD_Up) {
					dirOrder := elevio.ButtonEvent{
						Floor:  newFloor,
						Button: elevio.ButtonType(0),
					}
					completedOrders = append(completedOrders, dirOrder)
				} else if (masterOrderPanel[newFloor][1] != OT_NoOrder) & (localElevator.GetDirection() == MD_Down) {
					dirOrder := elevio.ButtonEvent{
						Floor:  newFloor,
						Button: elevio.ButtonType(1),
					}
					completedOrders = append(completedOrders, dirOrder)
				}

				takenOrders <- completedOrders
				SlaveState = Idle
			}

		case newButtons := <-drv_buttons:
			newOrders <- newButtons

		case priority := <-priOrderChan:
			if priority.Floor == -1 {
				SlaveState = Idle
			} else {
				SlaveState = Move
			}
		}
	}
}

func driveTo(localElevator *elevator.Elevator) {
	var lastFloor int = localElevator.getFloor()
	var newFloor int = localElevator.GetPriOrder().Floor

	if newFloor < lastFloor {
		elevio.SetMotorDirection(elevio.MD_Down)
		localElevator.setDirection(MD_Down)
	} else if newFloor < lastFloor {
		elevio.SetMotorDirection(elevio.MD_Up)
		localElevator.setDirection(MD_Up)
	}
}
