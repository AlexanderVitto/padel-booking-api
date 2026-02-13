package main

import (
	"log"

	"github.com/AlexanderVitto/padel-booking-api/internal/config"
	"github.com/AlexanderVitto/padel-booking-api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	srv := server.New(cfg)

	log.Printf("starting server on %s (env=%s)", cfg.Addr(), cfg.Env)
	if err := srv.Run(cfg.Addr()); err != nil {
		log.Fatal(err)
	}
}
