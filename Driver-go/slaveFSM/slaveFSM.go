package slaveFSM

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
)

func slaveFSMinit() {
	//KVA MÅ STOR FSM HA: 
		//LAGE ELEVATOR OBJEKTET!!! Med ID og sånt
		//

	//Make connection with local elevator, to make it run
	elevio.Init("localhost:15657", numFloors)

	elevio.SetMotorDirection(elevio.MD_Down)

	var doorOpen bool = false
	var moving bool = true
	var obs bool = false

	for f := 0; f < numFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}
}



func slaveFSM(id string, priOrderChan , orderMatrixChan) {

	//Channel where you get/update priorder, when you get it

	//channel 
	//go send update elns. som sender ut elevator struct heile tida

	//channel for nye ordre
	//go send nye ordre
		//Struct SlaveInformation
		//NewOrdres []elevio.ButtonEvent
		//CompletedOrders elevio.ButtonEvent


	drv_buttons := make(chan elevio.ButtonEvent) //For å hente nye knappetrykk
	drv_floors := make(chan int) //For å hente current floor
	drv_obs := make(chan bool) //for å hente obstruction

	//go ---.PollButtons(drv_buttons) 
	//go ---.PollFloorSensor(drv_floors)
	//go ---.PollObs(drv_obs)

	for {
		select {
		//Hente chn frå anna fil
		case priOrder := <-priOrderChan:
			//What to do if new pri order
			//Drive_to_() elns?
		//Hente chn frå anna fil
		case newOrderMatrix := <-orderMatrixChan:
			//oppdatere lys
		}
	}
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




