package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Davincible/goinsta"
	"github.com/fatih/color"
)

var (
	pathConfigModuleInstagram = pathConfigModules + string(os.PathSeparator) + "instagram.json"
	instagramConfig           configModuleInstagram
)

type configModuleInstagram struct {
	WaitMins int                        `json:"waitMins,omitempty"`
	Accounts []configModuleInstagramAcc `json:"accounts"`
}

type configModuleInstagramAcc struct {
	// Main
	Name         string   `json:"moduleName"`
	ID           string   `json:"id"`
	Destinations []string `json:"destinations"`

	WaitMins *int `json:"waitMins,omitempty"`
}

func loadConfig_Module_Instagram() error {
	prefixHere := "loadConfig_Module_Instagram(): "
	// TODO: Creation prompts if missing

	// LOAD JSON CONFIG
	if _, err := os.Stat(pathConfigModuleInstagram); err != nil {
		return fmt.Errorf("instagram config file not found: %s", err)
	} else {
		configBytes, err := os.ReadFile(pathConfigModuleInstagram)
		if err != nil {
			return fmt.Errorf("failed to read instagram config file: %s", err)
		} else {
			// Fix backslashes
			configStr := string(configBytes)
			configStr = strings.ReplaceAll(configStr, "\\", "\\\\")
			for strings.Contains(configStr, "\\\\\\") {
				configStr = strings.ReplaceAll(configStr, "\\\\\\", "\\\\")
			}
			// Parse
			if err = json.Unmarshal([]byte(configStr), &instagramConfig); err != nil {
				return fmt.Errorf("failed to parse instagram config file: %s", err)
			}
			// Output?
			if generalConfig.OutputSettings {
				s, err := json.MarshalIndent(instagramConfig, "", "\t")
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
	instagramEmail     string
	instagramPassword  string
	instagramConnected bool = false

	instagramScraper *goinsta.Instagram
)

func openInstagram() error {
	l := logInstructions{
		Location: "openInstagram",
		Task:     "login",
		Inline:   false,
		Color:    color.MagentaString,
	}

	if instagramEmail == "" || instagramPassword == "" {
		return errors.New("instagram credentials are incomplete")
	} else {
		log.Println(l.Log("Connecting to Instagram..."))

		//TODO: Proxy Support

		// Login Loop
		instagramLoginCount := 0
	do_instagram_login:
		instagramLoginCount++
		if instagramLoginCount > 1 {
			time.Sleep(3 * time.Second)
		}
		if instagramScraper, err := goinsta.Import(pathDataCookiesInstagram); err != nil {
			instagramScraper = goinsta.New(instagramEmail, instagramPassword)
			if err := instagramScraper.Login(); err != nil {
				log.Println(l.SetFlag(&lError).Log("Login Error: %s", err.Error()))
				if instagramLoginCount <= 3 {
					goto do_instagram_login
				} else {
					log.Println(l.SetFlag(&lError).Log("Failed to login to Instagram, the bot will not fetch this media..."))
					l.ClearFlag()
					return errors.New("login failed")
				}
			} else {
				log.Println(l.LogC(color.HiMagentaString, "Connected to %s via new login", instagramEmail))
				instagramConnected = true
				defer instagramScraper.Export(pathDataCookiesInstagram)
			}
		} else {
			log.Println(l.LogC(color.HiMagentaString, "Connected to %s via cache", instagramEmail))
			instagramConnected = true
		}
		//TODO: Reinforce Proxy Support
	}

	return nil
}

func handleInstagramAccount(account configModuleInstagramAcc) error {
	log.Printf(color.HiGreenString("<DEBUG> instagram account event fired: %s"), account.ID)
	return nil
}

/*func handleInstagramAccCmdOpts(config *configModuleInstagramAcc,
optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
s *discordgo.Session, i *discordgo.InteractionCreate) error {*/

func getInstagramAccConfigIndex(name string) int {
	for k, feed := range instagramConfig.Accounts {
		if strings.EqualFold(name, feed.Name) {
			return k
		}
	}
	return -1
}

func getInstagramAccConfig(name string) *configModuleInstagramAcc {
	i := getInstagramAccConfigIndex(name)
	if i == -1 {
		return nil
	} else {
		return &instagramConfig.Accounts[i]
	}
}

func existsInstagramConfig(name string) bool {
	for _, feed := range instagramConfig.Accounts {
		if strings.EqualFold(name, feed.Name) {
			return true
		}
	}
	return false
}

// func updateInstagramAccConfig(name string, config configModuleInstagramAcc) bool {

// func deleteInstagramAccConfig(name string) error
