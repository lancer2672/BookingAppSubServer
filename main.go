package main

import (
	"github.com/lancer2672/BookingAppSubServer/api"
	"github.com/lancer2672/BookingAppSubServer/db"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	serverConfig, err := utils.LoadConfig(".")
	if err != nil {
		log.Error().Err(err).Msg("Error loading config")
	}
	store := db.ConnectDatabase(serverConfig)
	api.NewServer(serverConfig, store)

}
