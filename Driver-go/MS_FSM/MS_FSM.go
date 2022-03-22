package MS_FSM

import (
	"Driver-go/elevator"
)

// case init
// case master
// case slave
// receive reports

//FUNC DOSTUFF(CHAN)

type SystemState int

const (
	Initialization SystemState = 0
	Connect                    = 1
	Slave                      = 2
	Master                     = 3
)

var myElevator elevator.Elevator
var orderPanel [elevator.NUMBER_OF_FLOORS][elevator.NUMBER_OF_COLUMNS]

func RunSystemFSM() {
	var sysState SystemState = Initialization
	for {
		switch sysState {
		case Initialization:
			//singleFSM-init
			//go runFSM
			//state = connect
		case Connect:
			//
		case Slave:

		case Master:

		}
	}
}
