package main

import (
	"Driver-go/elevio"
	//"Driver-go/orders"
)


type SlaveState int

const (
	Idle 		SlaveState = 0
	Move      		   	   = 1
	Obstruction            = 2
)


func ThaleSinMain() {
	slaveFSMinit()
	fmt.Println("Test")
	var masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var localElevator Elevator

	go slaveFSM(&localElevator, masterOrderPanel)
}


func SlaveFSMinit() {

	//Make connection with local elevator, to make it run
	elevio.Init("localhost:15657", NUMBER_OF_FLOORS)

	elevio.SetMotorDirection(elevio.MD_Down)

	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}


//MAIKEN HOPPER INN FOR DENNE:
func setLights(masterOrderPanel [orders.ConstNumFloors][orders.ConstNumElevators+2]int) {
	for f := 0; f < numFloors; f ++{
		for b := 0; b < 3; b++ {
			if ((b = 0)||(b = 1)) { //If up or down pushed
				elevio.SetButtonLamp(elevio.ButtonType(b), f, (masterOrderPanel[f][b]!=OT_NoOrder)) //Will set the lamp on/off if 0/1or2
			} else if (b = 2) { //If cab 
				elevio.SetButtonLamp(elevio.ButtonType(b), f, (masterOrderPanel[f][getElevatorIndex() + 2])!=OT_NoOrder)) //GetElevatorIndex gives the nr. of column
			}
		}
	}
}


//THALE JOBBER M RESTEN
func pollPriFloor() {
	for {
		priChan <- localElevator.GetPriOrder()
	}
}

func SlaveFSM(localElevator *elevator.Elevator, masterOrderPanel [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]int, takenOrders <-chan []elevio.ButtonEvent, newOrders <-chan []elevio.ButtonEvent) {
	
	var state SlaveState = Idle
	var currentDirection = MD_Down
	localElevator.setDirection = currentDirection

	//Og kjøre nedover til den når den nederste etasjen sin!

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	priOrderChan := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go pollPriFloor(priOrderChan) 
	go setLights(masterOrderPanel)

	for {		
		if (SlaveState = Move) {
			//Køyr til etasjen du skal til OG du må endre direction du går i (i localElevator), dersom du endrer denne!
			DriveTo(elevator.GetPriOrder(), &localElevator)
		} 

		select {
		case obstr := <-drv_obstr:
			switch {
			case SlaveState = Move:
				elevio.SetMotorDirection(elevio.MD_Stop)
				localElevator.setobs(true)
				SlaveState = Obstruction
			case SlaveState = Idle: 
				localElevator.setobs(true)
				SlaveState = Obstruction
			case SlaveState = Obstruction:
				if (localElevator.GetPriOrder() = OT_NoOrder) {
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

			if newFloor = localElevator.GetPriOrder().Floor {
				elevio.SetMotorDirection(elevio.MD_Stop)
				SetDoorOpenLamp(true)
				time.Sleep(3 * time.Second)
				SetDoorOpenLamp(false)
				cabOrder := elevio.ButtonEvent{
					Floor: newFloor, 
					Button: elevio.ButtonType(2),
				}
				if (localElevator.GetDirection() = MD_Down) {
					dirOrder := elevio.ButtonEvent{
						Floor: newFloor,
						Button:elevio.ButtonType(1),
					}
	
				} else if (localElevator.GetDirection() = MD_Up) {
					dirOrder := elevio.ButtonEvent{
						Floor: newFloor,
						Button:elevio.ButtonType(0),
					}
				}
				takenOrders <- [cabOrder, dirOrder]
				SlaveState = Idle
			}

		case newButtons := <-drv_buttons:
			newOrders <- newButtons

		case priority := <-priOrderChan:
			if (priority.Floor = -1) { 
				SlaveState = Idle
			} else {
				SlaveState = Move
			}
		}
	}
}

