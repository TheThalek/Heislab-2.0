package slaveFSM

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
)

const (
	Idle 		SlaveState = 0
	Move      		   = 1
	Obstruction            = 2
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
	// Sett start state lik noko?? Evt. ha ein default i state-machinen i SlaveFSM

	for f := 0; f < numFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}

func setLights(masterOrderPanel [ConstNumFloors][ConstNumElevators+2]int) {
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

func slaveFSM(localElevator *elevator.Elevator, masterOrderPanel [ConstNumFloors][ConstNumElevators+2]int) {
	
	setLights(masterOrderPanel)

	localElevator.SetObs(elevio.getObstruction()) //KANSKJE ENDRE SLIK AT DEN ER PEKER(?)

	if(elevio.getFloor() != -1) {
		localElevator.SetCurrentFloor(elevio.getFloor()) //Det samme som for SetObs, peker elns? Notasjon??
	}
	
	//update orders
	for f := 0; f < numFloors; f++ {
		for b := 0; b < numButtons; b++ {
			//If not already in matrix, 
			//New orders, in a list of button events?
		}
	}

	switch SlaveState {
	case Idle:
		if (priOrder != OT_NoOrder) {
			slaveState := Move
		}
	case Move: 

	case Obstruction:
		
	default: 
		slaveState := Idle
	}


	//Drive to PriOrder (Frå localElevator) / STATEMACHINE
}






//ENDRE DENNE!
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




