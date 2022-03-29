package main

import (
	"Driver-go/elevio"
)

const NUMBER_OF_FLOORS = 4
const NUMBER_OF_BUTTONS = 3
const NUMBER_OF_COLUMNS = NUMBER_OF_BUTTONS + 2
const NUMBER_OF_ELEVATORS = 3

type Elevator struct {
	direction    elevio.MotorDirection
	currentFloor int
	obs          bool
	priOrder     elevio.ButtonEvent
	index        string
}

func (e *Elevator) GetDirection() elevio.MotorDirection {
	return e.direction
}
func (e *Elevator) GetCurrentFloor() int {
	return e.currentFloor
}
func (e *Elevator) GetPriOrder() elevio.ButtonEvent {
	return e.priOrder
}
func (e *Elevator) GetObs() bool {
	return e.obs
}
func (e *Elevator) GetIndex() string {
	return e.index
}
func (e *Elevator) SetFloor(floor int) {
	e.currentFloor = floor
}
func (e *Elevator) SetDirection(dir elevio.MotorDirection) {
	e.direction = dir
}
func (e *Elevator) SetPriOrder(priOrder elevio.ButtonEvent) {
	e.priOrder = priOrder
}
func (e *Elevator) SetObs(obs bool) {
	e.obs = obs
}
func (e *Elevator) SetIndex(index string) {
	e.index = index
}

func (e *Elevator) DriveTo(order elevio.ButtonEvent) {

	var elevDir elevio.MotorDirection
	var motorDir elevio.MotorDirection

	if e.GetCurrentFloor() < order.Floor {
		motorDir = elevio.MD_Up
		elevDir = motorDir
	} else if e.GetCurrentFloor() > order.Floor {
		motorDir = elevio.MD_Down
		elevDir = motorDir
	} else {
		motorDir = elevio.MD_Stop
		if order.Button == elevio.BT_HallUp {
			elevDir = elevio.MD_Up
		} else if order.Button == elevio.BT_HallDown {
			elevDir = elevio.MD_Down
		}
	}

	e.SetDirection(elevDir)
	elevio.SetMotorDirection(motorDir)
}
