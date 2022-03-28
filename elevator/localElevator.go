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
	Init                   = 3
)

func ThaleSinMain() {
	LocalInit()
	var masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var myElevator Elevator
	startPri := elevio.ButtonEvent{
		Floor:  OT_NoOrder,
		Button: elevio.ButtonType(0),
	}
	myElevator.SetPriOrder(startPri)

	ElevIndexChan := make(chan int)
	takenOrdersChan := make(chan []elevio.ButtonEvent)
	newOrdersChan := make(chan elevio.ButtonEvent)

	go LocalControl(&myElevator, masterOrderPanel, takenOrdersChan, newOrdersChan, ElevIndexChan)
}

func LocalInit() {
	elevio.Init("localhost:15657", NUMBER_OF_FLOORS)
	elevio.SetMotorDirection(elevio.MD_Down)
	elevio.SetDoorOpenLamp(false)

	for f := 0; f < NUMBER_OF_FLOORS; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

//MAIKEN HOPPER INN FOR DENNE:
func setLights(masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, elevIndex int) {
	fmt.Println("B4 SetLights")
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
func pollPriFloor(priOrder chan elevio.ButtonEvent, myElevator *Elevator) {
	for {
		priOrder <- myElevator.GetPriOrder()
	}
}

func test(myElevator *Elevator) {
	fmt.Println("before 5 sec 1 ")
	time.Sleep(5 * time.Second)
	fmt.Println("after 5 sec 1")
	testPri1 := elevio.ButtonEvent{
		Floor:  3,
		Button: elevio.ButtonType(0),
	}
	myElevator.SetPriOrder(testPri1)

	fmt.Println("before 5 sec 2")
	time.Sleep(15 * time.Second)
	fmt.Println("after 5 sec 2")
	testPri2 := elevio.ButtonEvent{
		Floor:  1,
		Button: elevio.ButtonType(0),
	}
	myElevator.SetPriOrder(testPri2)
}

func FloorIntTest(elevIndex chan int) {
	for {
		elevIndex <- 1
	}
}

func LocalControl(myElevator *Elevator, masterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int, takenOrders chan []elevio.ButtonEvent, newOrders chan elevio.ButtonEvent, elevIndex chan int) {
	var state SlaveState = Init
	var currentDirection elevio.MotorDirection = elevio.MD_Down
	myElevator.SetDirection(currentDirection)
	//Og kjøre nedover til den når den nederste etasjen sin!

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	priOrderChan := make(chan elevio.ButtonEvent)

	//TESTER
	go test(myElevator)
	go FloorIntTest(elevIndex)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go pollPriFloor(priOrderChan, myElevator)

	//go setLights(masterOrderPanel, <-elevIndex) //TO DO; ha med en "polingsfunksjon" i main

	for {
		if state == Move {
			//Køyr til etasjen du skal til OG du må endre direction du går i (i myElevator), dersom du endrer denne!
			driveTo(myElevator, drv_floors)
		}

		select {
		case obstr := <-drv_obstr:
			switch {
			case state == Move:
				fmt.Println("drv_obstr Move")
				elevio.SetMotorDirection(elevio.MD_Stop)
				myElevator.SetObs(obstr)
				state = Obstruction
			case state == Idle:
				fmt.Println("drv_obstr Idle")
				myElevator.SetObs(obstr)
				state = Obstruction
			case state == Obstruction:
				fmt.Println("drv_obstr Obstruction")
				if myElevator.GetPriOrder().Floor == OT_NoOrder { //TO DO: All places where we se priOrder, we put priOrder to NoOrder if there isn't one!
					myElevator.SetObs(!obstr)
					state = Idle
				} else {
					myElevator.SetObs(!obstr)
					state = Move
				}
			}

		case newFloor := <-drv_floors:
			fmt.Println("drv_floors")
			myElevator.SetFloor(newFloor)
			elevio.SetFloorIndicator(newFloor)
			switch {
			case state == Init:
				fmt.Println("drv_floors Init")
				elevio.SetMotorDirection(elevio.MD_Stop)
				state = Idle
			case state == Move:
				fmt.Println("drv_floors Move")
				if newFloor == myElevator.GetPriOrder().Floor {
					fmt.Println("I have an order here!")
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					time.Sleep(3 * time.Second)
					elevio.SetDoorOpenLamp(false)

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
					fmt.Println("Completed order")
					takenOrders <- completedOrders
					fmt.Println("Sent to takenOrders")
					state = Idle
					fmt.Println("After setting equal to Idle")
				}
			}

		case newButtons := <-drv_buttons:
			fmt.Println("drv_buttons")
			newOrders <- newButtons

		case priority := <-priOrderChan:
			switch {
			case state != Init:
				fmt.Println("priOrderChan")
				if priority.Floor == OT_NoOrder {
					fmt.Println("priOrderChan Idle")
					state = Idle
				} else {
					fmt.Println("priOrderChan Move")
					state = Move
				}
			}
		}
	}
}

func driveTo(myElevator *Elevator, floorChan chan int) {
	fmt.Println("driveTo")
	var lastFloor int = myElevator.GetCurrentFloor()
	fmt.Println("lastFloor:", lastFloor)
	var newFloor int = myElevator.GetPriOrder().Floor
	fmt.Println("newFloor:", newFloor)
	if newFloor < lastFloor {
		fmt.Println("Going down")
		elevio.SetMotorDirection(elevio.MD_Down)
		myElevator.SetDirection(elevio.MD_Down)
	} else if newFloor > lastFloor {
		fmt.Println("Going up")
		elevio.SetMotorDirection(elevio.MD_Up)
		myElevator.SetDirection(elevio.MD_Up)
	} else if newFloor == lastFloor { //Hvis den er i rett etasje!
		fmt.Println("b4 Sending to drv_floors")
		floorChan <- newFloor
		fmt.Println("after Sending to drv_floors")
	}
}
