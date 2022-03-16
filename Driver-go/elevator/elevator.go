package elevator

import (
	"Driver-go/elevio"
)


type Elevator struct {
	direction    elevio.MotorDirection
	currentFloor int
	obs 		 bool
	priOrder	 elevio.ButtonEvent
	id			 string
}

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















// package elevator

// import (
// 	"Driver-go/elevio"
// )

// type Elevator struct {
// 	direction    elevio.MotorDirection
// 	currentFloor int
//	obs			 bool	 
// }

// func (e *Elevator) GetDirection() elevio.MotorDirection {
// 	return e.direction
// }
// func (e *Elevator) GetCurrentFloor() int {
// 	return e.currentFloor
// }
// func (e *Elevator) SetFloor(floor int) {
// 	e.currentFloor = floor
// }
// func (e *Elevator) SetDirection(dir elevio.MotorDirection) {
// 	e.direction = dir
// }

// func (e *Elevator) DriveTo(order elevio.ButtonEvent) {

// 	var elevDir elevio.MotorDirection
// 	var motorDir elevio.MotorDirection

// 	if e.GetCurrentFloor() < order.Floor {
// 		motorDir = elevio.MD_Up
// 		elevDir = motorDir
// 	} else if e.GetCurrentFloor() > order.Floor {
// 		motorDir = elevio.MD_Down
// 		elevDir = motorDir
// 	} else {
// 		motorDir = elevio.MD_Stop
// 		if order.Button == elevio.BT_HallUp {
// 			elevDir = elevio.MD_Up
// 		} else if order.Button == elevio.BT_HallDown {
// 			elevDir = elevio.MD_Down
// 		}
// 	}

// 	e.SetDirection(elevDir)
// 	elevio.SetMotorDirection(motorDir)
// }
