package singleFSM

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
	"Driver-go/orders"
)


type SlaveState int

const (
	Idle 		SlaveState = 0
	Move      		   	   = 1
	Obstruction            = 2
)


func ThaleSinMain() {
	slaveFSMinit()

	var masterOrderPanel [orders.ConstNumFloors][orders.ConstNumElevators+2]int

	var localElevator elevator.Elevator

	go slaveFSM(&localElevator, masterOrderPanel)

}


func slaveFSMinit() {

	//Make connection with local elevator, to make it run
	elevio.Init("localhost:15657", orders.ConstNumFloors)

	elevio.SetMotorDirection(elevio.MD_Down)

	// var doorOpen bool = false
	// var moving bool = true
	// var obs bool = false
	// Sett start state lik noko?? Evt. ha ein default i state-machinen i SlaveFSM

	for f := 0; f < orderrs.ConstNumFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}

	//Og kjøre nedover til den når den nederste etasjen sin!
	
}

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

func slaveFSM(localElevator *elevator.Elevator, masterOrderPanel [orders.ConstNumFloors][orders.ConstNumElevators+2]int) {
	//Skru lysene på/av ut ifrå masterOrderPanel som kontinuerlig blir tatt inn
	setLights(masterOrderPanel)
	
	drv_buttons := make(chan elevio.ButtonsEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	for {
		select {
		case obstr := <-drv_obstr:
			//Lagre obstruction i elevator_structen vår
			//localElevator.SetObs(elevio.getObstruction()) KANSKJE ENDRE SLIK AT DEN ER PEKER(?)

		case newfloor := <-drv_floors:
			fmt.Printf("%+v\n", a)
			//Oppdater etasjelys og elevator-objektet, slik at masterFSM veit kor du er

		case newButtons := <-drv_buttons:
			//Send informasjon om at knappen har blitt tatt, til masterFSM, 
		
		case newPriority := <-priOrderChan:
			driveTo(newPriority)
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
}

