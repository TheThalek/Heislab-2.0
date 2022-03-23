package main

//Nå funker ikke go run main.go
//Run alle filer: go run *.go
//Evt build and execute: go build && ./Elevator

//om du vil kjøre en og en, eller kun noen, go run .\main.go .\\Vår\Sanntidsprogrammerelevator.go .\orderLogic.go .\networking.go
func main() {
	//MS_FSM.maikenSinMain()
	//singleFSM.ThaleSinMain()
	go PederSinMain()
	//MS_FSM.RunSystemFSM()
	for {
	}
}
