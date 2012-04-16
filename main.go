package main

import (
	"io/ioutil"
	"log"

	"github.com/nictuku/chip-8/system"
)

func main() {
	sys := system.New()
	if err := sys.Init(); err != nil {
		log.Println(err)
		return
	}
	// Loadgame before.
	rom, err := ioutil.ReadFile("ibm.ch8")
	if err != nil {
		log.Println(err)
		return
	}
	sys.LoadGame(rom)
	sys.Run()
	log.Println("done")
}
