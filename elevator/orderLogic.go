package main

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

func PederSinOrderLogicMain() {
	var myElevator Elevator
	var elevatorPeers []Elevator
	elevatorPeers = append(elevatorPeers, myElevator)
	var MasterOrderPanel [NUMBER_OF_FLOORS][NUMBER_OF_COLUMNS]int

	var sysState SystemState = Initialization

	//hardware
	//network
	var id string
	var elevIndex int
	id = NetworkConnect("")

	msgTx := make(chan NetworkMessage)
	receivedMessages := make(chan NetworkMessage)
	roleChan := make(chan string)
	elevIndexChan := make(chan int)

	sysState = Connect
	go RunNetworkInterface(msgTx, receivedMessages, roleChan, elevIndexChan)
	for {
		select {
		case msg := <-receivedMessages:
			if sysState == Master {
				slaveInfo := ExtractSlaveInformation(msg)
			} else {
				masterInfo := ExtractMasterInformation(msg)
			}
		case role := <-roleChan:
			if role == MO_Master {
				sysState = Master
			} else if role == MO_Slave {
				sysState = Slave
			}

		case idx := <-indeelevIndexChan:
			elevIndex = idx
		default:
			switch sysState {
			case Master:

			case Slave:

			}
		}
	}
}
