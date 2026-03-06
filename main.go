package main

import (
	"log"

	"github.com/BaoPaper/SimpleNote/internal/notepad"
)

func main() {
	if err := notepad.Run(); err != nil {
		log.Fatal(err)
	}
}
