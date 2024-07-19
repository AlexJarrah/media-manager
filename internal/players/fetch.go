package players

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

func GetAllMediaPlayers(conn *dbus.Conn) (players []string, err error) {
	// List all names registered on the session bus
	var names []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return []string{}, err
	}

	// Filter names to include only MediaPlayer names
	for _, name := range names {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2.") {
			players = append(players, name)
		}
	}

	return players, nil
}

func GetPlayerByName(name string) (*Player, error) {
	for _, p := range *Players {
		if name == p.Name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("no player found with name: %s", name)
}

func GetPlayerBySignal(sender string, conn *dbus.Conn) (*Player, error) {
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	var name string
	err := obj.Call("org.freedesktop.DBus.GetNameOwner", 0, sender).Store(&name)
	if err != nil {
		return nil, fmt.Errorf("failed to get name owner: %v", err)
	}

	if name == "" {
		return nil, fmt.Errorf("no name found for sender %s", sender)
	}

	for _, p := range *Players {
		var owner string
		err = obj.Call("org.freedesktop.DBus.GetNameOwner", 0, p.Name).Store(&owner)
		if err == nil && owner == name {
			return p, nil
		}
	}

	players, err := GetAllMediaPlayers(conn)
	if err != nil {
		return nil, err
	}

	for _, p := range players {
		var owner string
		err = obj.Call("org.freedesktop.DBus.GetNameOwner", 0, p).Store(&owner)
		if err == nil && owner == name {
			player := &Player{Name: p}
			*Players = append(*Players, player)
			return player, nil
		}
	}

	return nil, fmt.Errorf("no MPRIS name found for sender %s", sender)
}

func getPlayerName(conn *dbus.Conn, sender string) (string, error) {
	obj := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus")
	var name string
	err := obj.Call("org.freedesktop.DBus.GetNameOwner", 0, sender).Store(&name)
	if err != nil {
		return "", fmt.Errorf("failed to get name owner: %v", err)
	}

	if name == "" {
		return "", fmt.Errorf("no name found for sender %s", sender)
	}

	// Now we need to get the well-known name for this unique name
	var names []string
	err = obj.Call("org.freedesktop.DBus.ListNames", 0).Store(&names)
	if err != nil {
		return "", fmt.Errorf("failed to list names: %v", err)
	}

	for _, n := range names {
		if strings.HasPrefix(n, "org.mpris.MediaPlayer2.") {
			var owner string
			err = obj.Call("org.freedesktop.DBus.GetNameOwner", 0, n).Store(&owner)
			if err == nil && owner == name {
				return n, nil
			}
		}
	}

	return "", fmt.Errorf("no MPRIS name found for sender %s", sender)
}
