package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

var (
	discordConfig configDiscordSettings
)

type configDiscordSettings struct {
	//LogLevel            int      `json:"apiLogLevel,omitempty"`
	//Timeout             int      `json:"timeout,omitempty"`
	//ExitOnBadConnection bool     `json:"exitOnBadConnection,omitempty"`
	//OutputMessages      bool     `json:"outputMessages,omitempty"`
	Admins              []string `json:"admins"`
	DeleteCommands      bool     `json:"deleteCommands"`
	PresenceEnabled     bool     `json:"presenceEnabled"`
	PresenceRefreshRate int      `json:"presenceRefreshRate"`
	PresenceType        string   `json:"presenceType,omitempty"`
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

	// FIX
	if discordConfig.PresenceRefreshRate == 0 {
		discordConfig.PresenceRefreshRate = 30
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

	log.Println(color.HiGreenString("Discord logged into \"%s\"#%s", discordUser.Username, discordUser.Discriminator))

	return nil
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
