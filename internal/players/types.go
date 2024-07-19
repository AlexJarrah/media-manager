package players

import "time"

type Player struct {
	Name               string        // Name of the player
	MBID               string        // MusicBrainz track ID
	StartListeningTime time.Time     // Time the track started playing
	TotalPlayTime      time.Duration // Total listening time; use GetTotalPlayTime() for accurate calculations
	IsPlaying          bool          // If the track is currently playing
	LastPlayStart      time.Time     // Time when track was last started/resumed
	Album              string        // Album title
	Artists            []string      // Track artists
	ArtURL             string        // Album art URL
	LengthSeconds      int64         // Track length in seconds
	Title              string        // Track Title
}
