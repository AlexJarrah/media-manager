package players

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

// Start/resume calculation of track play time
func (p *Player) Play() {
	if !p.IsPlaying {
		p.IsPlaying = true
		p.LastPlayStart = time.Now()
	}
}

// Pause calculation of track play time
func (p *Player) Pause() {
	if p.IsPlaying {
		p.IsPlaying = false
		p.TotalPlayTime += time.Since(p.LastPlayStart)
	}
}

// Reset track play time
func (p *Player) ResetPlayTime() {
	p.TotalPlayTime = 0
	p.LastPlayStart = time.Now()
}

// Calculate track play time
func (p *Player) GetTotalPlayTime() time.Duration {
	if p.IsPlaying {
		return p.TotalPlayTime + time.Since(p.LastPlayStart)
	}
	return p.TotalPlayTime
}

// Get player's current track position
func (p *Player) GetTrackPosition(conn *dbus.Conn) (float64, error) {
	obj := conn.Object(p.Name, "/org/mpris/MediaPlayer2")
	variant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Position")
	if err != nil {
		return 0, fmt.Errorf("failed to get Position property: %v", err)
	}

	// Position is in microseconds, convert to seconds
	positionMicroseconds, ok := variant.Value().(int64)
	if !ok {
		return 0, fmt.Errorf("Position is not in the expected format")
	}

	return float64(positionMicroseconds) / 1000000, nil
}
