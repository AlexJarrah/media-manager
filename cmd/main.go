package main

import (
	"log"

	"github.com/godbus/dbus/v5"

	"gitlab.com/AlexJarrah/media-manager/internal/database"
	"gitlab.com/AlexJarrah/media-manager/internal/filesystem"
	"gitlab.com/AlexJarrah/media-manager/internal/monitor"
)

func main() {
	if err := filesystem.Initialize(); err != nil {
		log.Fatal(err)
	}

	if err := database.Initialize(); err != nil {
		log.Fatal(err)
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err = monitor.MonitorPlayers(conn); err != nil {
		log.Fatal(err)
	}
}
