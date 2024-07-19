# Media Manager

Media Manager seamlessly integrates your media players with services like Last.fm and Discord, with more integrations planned. Enjoy automatic scrobbling and real-time Discord presence updates for local music files, connected phones, music streaming services, etc.

## Features

- **Whitelisted Players**: Only monitor specified media players.
- **Scrobbling**: Automatically log your music listening history to Last.fm.
- **Discord Rich Presence**: Sync your Discord status with your current track.

## Usage

App configuration is stored in a JSON file located at: `~/.config/media-manager/config.json`, if you don't see this file, run the app for the first time to generate it.

In this file, you can set your preferences & configuration. Once you fill out all fields, simply run the app to start.

You can use the below command to list all available players to determine player names:

```bash
dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep 'string "org.mpris.' | awk -F'"' '{print $2}'
```

## Building

```bash
curl -fsSL https://gitlab.com/AlexJarrah/media-manager/-/raw/main/scripts/build.sh | sh
./media-manager
```

## License

This project is licensed under the MIT License - see the [LICENSE](https://gitlab.com/AlexJarrah/media-manager/-/raw/main/LICENSE) file for details.
