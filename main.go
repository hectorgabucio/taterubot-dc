package main

import (
	"github.com/hectorgabucio/taterubot-dc/bootstrap"
	"log"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
