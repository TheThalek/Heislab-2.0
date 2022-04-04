package main

import (
	"Driver-go/elevio"
	//"fmt"
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
	index        int
	online       bool
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
func (e *Elevator) GetIndex() int {
	return e.index
}
func (e *Elevator) GetOnline() bool {
	return e.online
}
func (e *Elevator) SetFloor(floor int) {
	e.currentFloor = floor
}
func (e *Elevator) SetDirection(dir elevio.MotorDirection) {
	if dir != 0 {
		e.direction = dir
	}
}
func (e *Elevator) SetPriOrder(priOrder elevio.ButtonEvent) {
	e.priOrder = priOrder
}
func (e *Elevator) SetObs(obs bool) {
	e.obs = obs
}
func (e *Elevator) SetIndex(index int) {
	e.index = index
}
func (e *Elevator) SetOnline(online bool) {
	e.online = online
}

func NewElevator() Elevator {
	return Elevator{
		direction:    elevio.MD_Stop,
		currentFloor: -1,
		obs:          false,
		priOrder:     elevio.ButtonEvent{Floor: -1, Button: elevio.BT_Cab},
		index:        -1,
		online:       false,
	}
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
