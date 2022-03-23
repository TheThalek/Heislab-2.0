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

// func setLights(masterOrderPanel [orders.ConstNumFloors][orders.ConstNumElevators+2]int) {
// 	for f := 0; f < numFloors; f ++{
// 		for b := 0; b < 3; b++ {
// 			if ((b = 0)||(b = 1)) { //If up or down pushed
// 				elevio.SetButtonLamp(elevio.ButtonType(b), f, (masterOrderPanel[f][b]!=OT_NoOrder)) //Will set the lamp on/off if 0/1or2
// 			} else if (b = 2) { //If cab 
// 				elevio.SetButtonLamp(elevio.ButtonType(b), f, (masterOrderPanel[f][getElevatorIndex() + 2])!=OT_NoOrder)) //GetElevatorIndex gives the nr. of column
// 			}
// 		}
// 	}
// }

func pollPriority(localElevator *Elevator, priChan chan elevio.ButtonEvent) {
	for {
		priChan <- localElevator.GetPriOrder()
	}
}

func SlaveFSM(localElevator *Elevator, masterOrderPanel [orders.ConstNumFloors][orders.ConstNumElevators+2]int) {
	//Starter i Idle
	var state SlaveState = Idle

	//Og kjøre nedover til den når den nederste etasjen sin!

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	//Skal lages i stor FSM
	// taken_orders := make(chan []elevio.ButtonEvent)
	// new_orders := make(chan []elevio.ButtonEvent)


	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go pollPriority(priOrderChan)


	for {
		//Skru lysene på/av ut ifrå masterOrderPanel som kontinuerlig blir tatt inn
		setLights(masterOrderPanel)
		//driveTo
		select {
		case obstr := <-drv_obstr:
			//Hvis obs skrus på: 
				//Lagre obstruction i elevator_structen vår
				//localElevator.SetObs(elevio.getObstruction()) KANSKJE ENDRE SLIK AT DEN ER PEKER(?)

				//Ikkje køyr vidare til pri_floor, motordirection=stop!
				//Gå til IDLE

			//Hvis obs skrus av: 
				//Lagre obstruction i elevator_structen vår
				//localElevator.SetObs(elevio.getObstruction()) KANSKJE ENDRE SLIK AT DEN ER PEKER(?)

				//Køyr vidare, hvis vi har 
				//

		case newfloor := <-drv_floors:
			fmt.Printf("%+v\n", a)
			//Oppdater etasjelys og elevator-objektet, slik at masterFSM veit kor du er
			//Eneste stedet du har lov til å stoppe
			//Sjekke om du er i samme etasjen som prifloor er!
				//Stopp
				//Åpne dørene i 3 sek
				//Lukk dørene og gå videre til anna state (?) 
				//Send ut på kanalen om at relevante ordre har blitt tatt.
					//Sjekk om 

		case newButtons := <-drv_buttons:
			//Send informasjon om at knappen har blitt tatt, til masterFSM, 
			//Har ikkje noko å seie for heisen ellers, skal berre informere stor FSM om dette
			//Skriver dette til kanalen

		case newPriority := <-priOrderChan:
			//Ved ny priOrder, finn ut kor du er, køyr til priOrder(?)


		}

	}
}

