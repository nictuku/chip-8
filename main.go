package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/nictuku/chip-8/system"
)

func main() {
	flag.Parse()
	romPath := flag.Arg(0)
	if romPath == "" {
		log.Fatal("Missing ROM filename argument.")
	}
	log.Printf("Loading ROM from %v", romPath)
	rom, err := ioutil.ReadFile(romPath)
	if err != nil {
		log.Fatal(err)
	}

	sys := system.New()
	if err := sys.Init(); err != nil {
		log.Println(err)
		return
	}
	sys.LoadGame(rom)
	if err := sys.Run(); err != nil {
		log.Fatal(err)
	}
	log.Println("done")
}
