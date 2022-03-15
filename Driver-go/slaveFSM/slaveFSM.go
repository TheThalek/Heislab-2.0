package slaveFSM

import (
	"Driver-go/elevator"
	"Driver-go/elevio"
	"Driver-go/orders"
)

//Struct
type Elevator struct {
	direction    elevio.MotorDirection
	currentFloor int
	priFloor	 elevio.ButtonEvent
	obstruction	 bool
}

//GET OG SET FUNC For structen
func (e *Elevator) GetDirection() elevio.MotorDirection {
	return e.direction
}
func (e *Elevator) GetCurrentFloor() int {
	return e.currentFloor
}
func (e *Elevator) GetPriFloor() elevio.ButtonEvent {
	return e.priFloor
}
func (e *Elevator) GetObstruction() bool {
	return e.obstruction
}
func (e *Elevator) SetFloor(floor int) {
	e.currentFloor = floor
}
func (e *Elevator) SetDirection(dir elevio.MotorDirection) {
	e.direction = dir
}
func (e *Elevator) SetPriFloor(priFloor elevio.ButtonEvent) {
	e.priFloor = priFloor
}
func (e *Elevator) SetObstruction(obs bool) {
	e.obstruction = obs
}




func slaveFSM(elevio.ButtonEvent) {
	//INIT:
	numFloors := 4

	//Make the elevator-object/struct
	var myElevator elevator.Elevator

	//Make connection with local elevator, to make it run
	elevio.Init("localhost:15657", numFloors)

	//Get buttons pressed on local order panel

	//Send buttons pressed on local order panel

	//Get updated order-panel
	var orderPanel [orders.ConstNumFloors][3]int

	elevio.SetMotorDirection(elevio.MD_Down)

	var doorOpen bool = false
	var moving bool = true
	var obs bool = false


	//Ta inn: priOrder
	//Sende ut: på en kanal(?) posisjon, 
	//om den har tatt en ordre/er ferdig, 
	//direction, obstruksjon, 



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




