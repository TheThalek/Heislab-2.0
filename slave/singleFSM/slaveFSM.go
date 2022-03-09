package slaveFSM

import (
	"slave/elevator"
	"slave/elevator.io"
)

numFloors := 4 //To do; Kanskje gjør denne "global", då den blir brukt av alle?


func init() {
	var myElevator elevator.Elevator

	var orderPanel [orders.ConstNumFloors][3]int

	elevio.Init("localhost:15657", numFloors)

	elevFSM.RunElevFSM(numFloors, myElevator, orderPanel)
	
	elevio.SetMotorDirection(elevio.MD_Down)

	var doorOpen bool = false
	var moving bool = true
	var obs bool = false
}

//Tar inn ein Elevator-struct, som 
func slaveFSM(elevio.ButtonEvent ) {
	//Ta inn: priOrder
	//Sende ut: på en kanal(?) posisjon, 
	//om den har tatt en ordre/er ferdig, 
	//direction, obstruksjon, 

	init()

	var priorityOrder elevio.ButtonEvent
	priorityOrder.Floor = -1

	//Turns on all lights, but have to do this from the order matrix -> change this one. 
	for f := 0; f < numFloors; f++ {
		for b := 0; b < 3; b++ {
			elevio.SetButtonLamp(elevio.ButtonType(b), f, false)
		}
	}

	//Channel where you get/update priorder, when you get it

}


func main() {

}




