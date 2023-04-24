package main

import "github.com/bwmarrin/discordgo"

const (
	projectName        = "DISCORD FEEDBOT (DFB)"
	projectVersion     = "1.0.0-a.0"
	projectNameVersion = projectName + " " + projectVersion

	projectRepo          = "get-got/discord-feedbot"
	projectRepoURL       = "https://github.com/" + projectRepo
	projectReleaseURL    = projectRepoURL + "/releases/latest"
	projectReleaseApiURL = "https://api.github.com/repos/" + projectRepo + "/releases/latest"
)

const (
	pathData = "data"
)

func getComponentVersions() map[string]string {
	return map[string]string{
		projectName:     projectVersion,
		"discordgo":     "v" + discordgo.VERSION,
		"Discord API":   "v" + discordgo.APIVersion,
		"Twitter API":   "v1.1",
		"Instagram API": "vX",
	}
}
