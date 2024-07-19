package players

import "github.com/godbus/dbus/v5"

// Update the player's track metadata
func (p *Player) UpdateMediaPlayerMetadata(variant dbus.Variant, conn *dbus.Conn) {
	metadata := variant.Value().(map[string]dbus.Variant)
	for key, value := range metadata {
		v := value.Value()
		switch key {
		case "xesam:album":
			p.Album = v.(string)
		case "xesam:artist":
			p.Artists = v.([]string)
		case "mpris:artUrl":
			p.ArtURL = v.(string)
		case "mpris:length":
			p.LengthSeconds = v.(int64) / 1000000
		case "xesam:title":
			p.Title = v.(string)
		}
	}
}
