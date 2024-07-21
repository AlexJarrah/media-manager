package monitor

import (
	"fmt"
	"log"
	"time"

	"github.com/godbus/dbus/v5"

	"gitlab.com/AlexJarrah/media-manager/internal/database"
	"gitlab.com/AlexJarrah/media-manager/internal/filesystem"
	"gitlab.com/AlexJarrah/media-manager/internal/players"
	"gitlab.com/AlexJarrah/media-manager/internal/providers/lastfm"
	"gitlab.com/AlexJarrah/media-manager/internal/providers/musicbrainz"
)

func MonitorPlayers(conn *dbus.Conn) error {
	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)

	err := conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/mpris/MediaPlayer2"),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	)
	if err != nil {
		return fmt.Errorf("failed to add match rule: %v", err)
	}

	log.Println("Monitoring for track changes...")
	handleSignals(c, conn)
	return nil
}

func handleSignals(c <-chan *dbus.Signal, conn *dbus.Conn) {
	for sig := range c {
		player, props, ok := validateSignal(sig, conn)
		if !ok {
			continue
		}

		if variant, ok := props["Metadata"]; ok {
			handleNewTrack(player, variant, conn)
		}

		if status, ok := props["PlaybackStatus"]; ok {
			handlePlaybackStatus(player, status)
		}
	}
}

func validateSignal(sig *dbus.Signal, conn *dbus.Conn) (*players.Player, map[string]dbus.Variant, bool) {
	player, err := players.GetPlayerBySignal(sig.Sender, conn)
	if err != nil {
		log.Printf("Error getting player name: %v\n", err)
		return nil, nil, false
	}

	if !player.IsWhitelisted() {
		log.Println("Ignored player:", player.Name)
		return nil, nil, false
	}

	ifaceName, ok := sig.Body[0].(string)
	if !ok || ifaceName != "org.mpris.MediaPlayer2.Player" {
		return nil, nil, false
	}

	props, ok := sig.Body[1].(map[string]dbus.Variant)
	if !ok {
		return nil, nil, false
	}

	return player, props, true
}

func handleNewTrack(player *players.Player, variant dbus.Variant, conn *dbus.Conn) {
	if player.Title != "" {
		go onTrackChange(*player)
	}

	player.ResetPlayTime()
	player.UpdateMediaPlayerMetadata(variant, conn)
	player.StartListeningTime = time.Now()
}

func handlePlaybackStatus(player *players.Player, status dbus.Variant) {
	playbackStatus, _ := status.Value().(string)
	switch playbackStatus {
	case "Playing":
		player.Play()
	case "Stopped":
		player.Pause()
	}
	log.Printf("Playback status changed: %s (%s)", playbackStatus, player.Title)
}

func onTrackChange(player players.Player) {
	log.Printf("Total track play time: %v (%s)\n", player.GetTotalPlayTime(), player.Title)

	db, err := database.NewDB()
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	tracks, err := db.GetTracks("name", fmt.Sprintf("%s", player.Title))
	if err != nil {
		log.Println(err)
		return
	}

	if len(tracks) == 0 {
		log.Println("track not found in database")
		return
	}

	track := tracks[0]
	listen := database.Listen{
		UserID:     1,
		TrackID:    track.ID,
		ListenTime: int(player.GetTotalPlayTime().Seconds()),
		Timestamp:  time.Now(),
	}

	err = db.AddListens([]*database.Listen{&listen})
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Logged track listen to %s (%.0fs)", player.Title, player.GetTotalPlayTime().Seconds())

	// Scrobble only if 50%+ of the track was listened to
	listenedPercentage := (player.GetTotalPlayTime().Seconds() / float64(player.LengthSeconds)) * 100
	if listenedPercentage > 50 {
		var err error
		player.MBID, err = musicbrainz.FetchMBID(player)
		if err != nil {
			log.Println(err)
		}

		config, err := filesystem.GetConfigFile()
		if err != nil {
			log.Println(err)
			return
		}

		sessionKey, err := lastfm.Authenticate(config.LastFM)
		if err != nil {
			log.Println(err)
			return
		}

		err = lastfm.Scrobble(player, config.LastFM.APIKey, config.LastFM.APISecret, sessionKey)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		log.Printf("Track listening time (%.0fs) is less than 50%% (%.0f%%) of the track length (%ds). Not scrobbling %s.\n", player.GetTotalPlayTime().Seconds(), listenedPercentage, player.LengthSeconds, player.Title)
	}
}
