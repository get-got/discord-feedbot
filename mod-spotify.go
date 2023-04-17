package main

import "os"

var (
	pathConfigModuleSpotify = pathConfigModules + string(os.PathSeparator) + "spotify.json"
)

var (
	spotifyClientID     string
	spotifyClientSecret string
)
