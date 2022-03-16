package slaveFSM

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
)

func slaveFSMinit(int numFloors) {
	//KVA MÅ STOR FSM HA: 
		//LAGE ELEVATOR OBJEKTET!!! Med ID og sånt
		//

	//Make connection with local elevator, to make it run
	elevio.Init("localhost:15657", numFloors)

	elevio.SetMotorDirection(elevio.MD_Down)

	// var doorOpen bool = false
	// var moving bool = true
	// var obs bool = false

	for f := 0; f < numFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}


func slaveFSM(localElevator *elevator.Elevator, masterOrderPanel [ConstNumFloors][ConstNumElevators+2]int) {
	//Oppdater lysene ut ifrå masterOrderPanel-kopien (Både skru av(0) og på(1/2))
	for f := 0; f <= 2; f ++{
		for b := 0; b < 3; b++ {
			if(f = 0 or f = 1){ //If up or down pushed
				
			}
			if(f = 3){ //If cab 

			}
		}
	}



	for f := 0; f < numFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}



	//Oppdater localElevator
		//direction
		//current floor
		//obs
	//Drive to PriOrder (Frå localElevator)
	//

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




