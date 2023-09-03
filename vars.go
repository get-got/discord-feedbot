package main

import "os"

const (
	projectName        = "discord-feedbot"
	projectLabel       = "DISCORD FEED BOT (DFB)"
	projectVersion     = "1.0.0-a.230903"
	projectNameVersion = projectName + " " + projectVersion
	projectColor       = "#00FFFF"

	projectRepo          = "get-got/discord-feedbot"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"
)

var (
	pathData = "data"

	pathDataCookies        = pathData + string(os.PathSeparator) + "cookies"
	pathDataCookiesTwitter = pathDataCookies + string(os.PathSeparator) + "twitter.json"
)
