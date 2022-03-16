package slaveFSM

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
)



func slaveFSM(numFloors int, myElevator elevator.Elevator, orderPanel [orders.ConstNumFloors][3]int) {
	//INIT:
	const numFloors int = 4

	//Make the elevator-object/struct
	var myElevator elevator.Elevator

	//Make connection with local elevator, to make it run
	elevio.Init("localhost:15657", numFloors)

	//Get buttons pressed on local order panel

	//Send buttons pressed on local order panel

	//Get updated order-panel
	var localPanel [numFloors][3]int

	elevio.SetMotorDirection(elevio.MD_Down)

	var doorOpen bool = false
	var moving bool = true
	var obs bool = false

	for f := 0; f < numFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}

	//Ta inn: priOrder
	//Sende ut: på en kanal(?) posisjon, 
	//om den har tatt en ordre/er ferdig, 
	//direction, obstruksjon, 

	//Channel where you get/update priorder, when you get it

}





func (e *Elevator) DriveTo(priOrder elevio.ButtonEvent) {
	var elevDir elevio.MotorDirection
	var motorDir elevio.MotorDirection

	if e.GetCurrentFloor() < priOrder.floor {
		motorDir = elevio.MD_up
		elevdir = motorDi
	} else if e.GetCurrentFloor() > priOrder.Floor {
		motorDir = elevio.MD_Down
		elevDir = motorDir
	} else {
		motorDir = elevio.MD_Stop
		if priOrder.Button == elevio.BT_HallUp {
			elevDir = elevio.MD_Up
		} else if priOrder.Button == elevio.BT_HallDown {
			elevDir = elevio.MD_Down
		}
	}

	e.SetDirection(elevDir)
	elevio.SetMotorDirection(motorDir)
}




