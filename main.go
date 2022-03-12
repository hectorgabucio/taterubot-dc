package main

import (
	"github.com/hectorgabucio/taterubot-dc/bootstrap"
	"log"
)

//go:generate go-localize -input localizations_src -output localizations

func main() {
	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}
