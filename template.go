// Package Dispatcher
package template

func Dispatcher() {
	// receiveMessage("symfony_request")
	forever := make(chan bool)
	go receiveMessage("symfony_request")
	go receiveMessage("sbom_dispatcher")
	<-forever
}

func main() {
	Dispatcher()
}
