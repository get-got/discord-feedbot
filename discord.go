package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"gopkg.in/ini.v1"
)

var (
	discordConfig configDiscordSettings

	discordConfigDef_Presence_Enabled bool = true
)

type configDiscordPresence struct {
	Enabled       *bool                  `json:"enabled"`       // really just to optionally disable
	Type          string                 `json:"type"`          // Online, Idle, DND, Invisible
	Label         discordgo.ActivityType `json:"label"`         // Playing[0], Streaming[1], Listening[2], Watching[3], Custom[4,DOESNT WORK], Competing[5]
	Status        string                 `json:"status"`        // text
	StatusDetails string                 `json:"statusDetails"` // text
	Duration      int                    `json:"duration"`      // seconds to sleep after changing to this.
}

type configDiscordSettings struct {
	//LogLevel            int      `json:"apiLogLevel,omitempty"`
	//Timeout             int      `json:"timeout,omitempty"`
	//ExitOnBadConnection bool     `json:"exitOnBadConnection,omitempty"`
	//OutputMessages      bool     `json:"outputMessages,omitempty"`
	Admins         []string                `json:"admins"`
	DeleteCommands bool                    `json:"deleteCommands"`
	Presence       []configDiscordPresence `json:"presence"`
}

var discordConfigDefault = configDiscordSettings{
	Presence: []configDiscordPresence{
		{
			Enabled:       &discordConfigDef_Presence_Enabled,
			Type:          string(discordgo.StatusOnline),
			Label:         discordgo.ActivityTypeGame,
			Status:        "DFB {{dfbVersion}}",
			StatusDetails: "<< STATUS {{presenceCount}} @ {{presenceDuration}} >>",
			Duration:      15,
		},
		{
			Enabled:       &discordConfigDef_Presence_Enabled,
			Type:          string(discordgo.StatusDoNotDisturb),
			Label:         discordgo.ActivityTypeWatching,
			Status:        "{{linkCount}} updates sent",
			StatusDetails: "<< STATUS {{presenceCount}} @ {{presenceDuration}} >>",
			Duration:      30,
		},
		{
			Enabled:       &discordConfigDef_Presence_Enabled,
			Type:          string(discordgo.StatusDoNotDisturb),
			Label:         discordgo.ActivityTypeWatching,
			Status:        "{{feedCount}}",
			StatusDetails: "<< STATUS {{presenceCount}} @ {{presenceDuration}} >>",
			Duration:      30,
		},
		{
			Enabled:       &discordConfigDef_Presence_Enabled,
			Type:          string(discordgo.StatusDoNotDisturb),
			Label:         discordgo.ActivityTypeListening,
			Status:        "{{numServers}} servers",
			StatusDetails: "<< STATUS {{presenceCount}} @ {{presenceDuration}} >>",
			Duration:      15,
		},
		{
			Enabled:       &discordConfigDef_Presence_Enabled,
			Type:          string(discordgo.StatusIdle),
			Label:         discordgo.ActivityTypeWatching,
			Status:        "for {{uptime}}",
			StatusDetails: "<< STATUS {{presenceCount}} @ {{presenceDuration}} >>",
			Duration:      30,
		},
	},
}

func loadConfig_Discord() error {
	prefixHere := "loadConfig_Discord(): "
	// TODO: Creation prompts if missing

	// LOAD INI CREDS
	if _, err := os.Stat(pathConfigDiscordCredentials); err == nil {
		config, err := ini.Load(pathConfigDiscordCredentials)
		if err != nil {
			return fmt.Errorf("failed to parse discord credentials file: %s", err)
		} else {
			discordToken = config.Section("").Key("token").String()
			if len(discordToken) < 50 {
				return errors.New("discord token length is too short")
			}
		}
	} else {
		return fmt.Errorf("discord credentials file not found: %s", err)
	}

	// LOAD JSON CONFIG
	if _, err := os.Stat(pathConfigDiscordSettings); err != nil {
		return fmt.Errorf("discord config file not found: %s", err)
	} else {
		configBytes, err := os.ReadFile(pathConfigDiscordSettings)
		if err != nil {
			return fmt.Errorf("failed to read discord config file: %s", err)
		} else {
			// Fix backslashes
			configStr := string(configBytes)
			configStr = strings.ReplaceAll(configStr, "\\", "\\\\")
			for strings.Contains(configStr, "\\\\\\") {
				configStr = strings.ReplaceAll(configStr, "\\\\\\", "\\\\")
			}
			// Parse
			discordConfig = discordConfigDefault
			if err = json.Unmarshal([]byte(configStr), &discordConfig); err != nil {
				return fmt.Errorf("failed to parse discord config file: %s", err)
			}
			// Output?
			if generalConfig.OutputSettings {
				s, err := json.MarshalIndent(discordConfig, "", "\t")
				if err != nil {
					log.Println(color.HiRedString(prefixHere+"failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString(prefixHere+"\n%s", color.YellowString(string(s))))
				}
			}
		}
	}

	return nil
}

var (
	discordToken string

	discord     *discordgo.Session
	discordUser *discordgo.User
)

func openDiscord() error {
	var err error

	l := logInstructions{
		Location: "openDiscord()",
		Task:     "login",
		Inline:   false,
		Color:    color.GreenString,
	}

	discord, err = discordgo.New("Bot " + discordToken)
	if err != nil {
		return fmt.Errorf("error creating discord client: %s", err.Error())
	}

	err = discord.Open()
	if err != nil {
		return fmt.Errorf("error logging into discord client: %s", err.Error())
	}

	discord.State.MaxMessageCount = 100000
	discord.State.TrackChannels = true
	discord.State.TrackThreads = true
	discord.State.TrackMembers = true
	discord.State.TrackThreadMembers = true

	discordUser, err = discord.User("@me")
	if err != nil {
		discordUser = discord.State.User
		if discordUser == nil {
			return errors.New("error checking discord client")
		}
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	log.Println(l.LogCI(lSuccess.Color, true, "Discord logged into \"%s\"#%s", discordUser.Username, discordUser.Discriminator))

	return nil
}

func runDiscordPresences() {
	// Rotate Presences
	for presenceKey, presence := range discordConfig.Presence {
		enabled := false
		if presence.Enabled == nil {
			enabled = true
		} else {
			enabled = *presence.Enabled
		}
		if enabled {
			if presence.Duration == 0 {
				presence.Duration = 15
			}
			// Only change status type, no text.
			if presence.Status == "" {
				discord.UpdateStatusComplex(discordgo.UpdateStatusData{
					Status: presence.Type,
				})
			} else {
				// Format state (referring to it as details) - Presence-specific key replacements
				dataKeyReplacementPresence := func(input string) string {
					input = dataKeyReplacement(input)
					if strings.Contains(input, "{{presenceCount}}") {
						input = strings.ReplaceAll(input, "{{presenceCount}}",
							fmt.Sprintf("%d/%d", presenceKey+1, len(discordConfig.Presence)))
					}
					if strings.Contains(input, "{{presenceDuration}}") {
						input = strings.ReplaceAll(input, "{{presenceDuration}}",
							shortenTime(durafmt.ParseShort(
								time.Duration(presence.Duration*int(time.Second)),
							).String()),
						)
					}
					return input
				}
				// Update
				discord.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name:  dataKeyReplacementPresence(presence.Status),
						State: dataKeyReplacementPresence(presence.StatusDetails),
						Type:  discordgo.ActivityType(presence.Label), // Playing/Listening/Watching/etc
					}},
					Status: presence.Type, // online/idle/dnd/invisible
				})
			}
			time.Sleep(time.Duration(presence.Duration * int(time.Second)))
		}
	}
}

func isBotAdmin(id string) bool {
	for _, admin := range discordConfig.Admins {
		if admin == id {
			return true
		}
	}
	return false
}

func getAuthor(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member == nil && i.Message.Author == nil {
		return nil
	}
	if i.Member != nil {
		return i.Member.User
	} else if i.Message.Author != nil {
		return i.Message.Author
	}
	return nil
}
