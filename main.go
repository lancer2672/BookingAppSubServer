package main

import (
	"os"

	"github.com/lancer2672/BookingAppSubServer/api"
	"github.com/lancer2672/BookingAppSubServer/db"
	"github.com/lancer2672/BookingAppSubServer/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	serverConfig, err := utils.LoadConfig(".")
	if err != nil {
		log.Error().Err(err).Msg("Error loading config")
		os.Exit(-1)
	}
	store := db.ConnectDatabase(serverConfig)
	server, err := api.NewServer(serverConfig, store)
	if err != nil {
		log.Error().Err(err).Msg("Error loading config")
		os.Exit(-1)
	}
	server.Start(serverConfig.ServerAddress)
}
