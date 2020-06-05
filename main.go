package main

import "github.com/bradford-hamilton/apple-1/internal/vm"

func main() {
	vm := vm.New()
	go vm.Run()
	<-vm.ShutdownC
}
