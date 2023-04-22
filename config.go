package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/ini.v1"
)

/*

.ini		General Settings, Discord Settings, and Module Credentials
.json		Module Settings

*/

//#region Base Vars / Loading

var (
	pathConfig = "config"
)

func loadConfig() []error {
	var errors []error
	var tmperr error

	if tmperr = loadConfig_General(); tmperr != nil {
		errors = append(errors, tmperr)
	}

	if tmperr = loadConfig_Discord(); tmperr != nil {
		errors = append(errors, tmperr)
	}

	if tmperr = loadConfig_Modules_Credentials(); tmperr != nil {
		errors = append(errors, tmperr)
	}

	tmperrs := loadConfig_Modules()
	for _, err := range tmperrs {
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

//#endregion

//#region General

var (
	pathConfigGeneralSettings    = pathConfig + string(os.PathSeparator) + "general.json"
	pathConfigDiscordCredentials = pathConfig + string(os.PathSeparator) + "discord.ini"
	pathConfigDiscordSettings    = pathConfig + string(os.PathSeparator) + "discord.json"
)

func loadConfig_General() error {
	// TODO: Creation prompts if missing

	// LOAD JSON CONFIG
	if _, err := os.Stat(pathConfigGeneralSettings); err != nil {
		return fmt.Errorf("general config file not found: %s", err)
	} else {
		configBytes, err := os.ReadFile(pathConfigGeneralSettings)
		if err != nil {
			return fmt.Errorf("failed to read general config file: %s", err)
		} else {
			// Fix backslashes
			configStr := string(configBytes)
			configStr = strings.ReplaceAll(configStr, "\\", "\\\\")
			for strings.Contains(configStr, "\\\\\\") {
				configStr = strings.ReplaceAll(configStr, "\\\\\\", "\\\\")
			}
			// Parse
			if err = json.Unmarshal([]byte(configStr), &generalConfig); err != nil {
				return fmt.Errorf("failed to parse general config file: %s", err)
			}
			// Output?
			if generalConfig.OutputSettings {
				s, err := json.MarshalIndent(generalConfig, "", "\t")
				if err != nil {
					log.Println(color.HiRedString("failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString("loadConfig_General():\n%s", color.YellowString(string(s))))
				}
			}
		}
	}

	return nil
}

var (
	generalConfig configGeneralSettings
)

type configGeneralSettings struct {
	Debug          bool `json:"debug,omitempty"`
	LogLevel       int  `json:"logLevel,omitempty"`
	OutputSettings bool `json:"outputSettings,omitempty"`
}

//#endregion

//#region Modules

var (
	pathConfigModules            = pathConfig + string(os.PathSeparator) + "modules"
	pathConfigModulesCredentials = pathConfigModules + string(os.PathSeparator) + "credentials.ini"
)

func loadConfig_Modules_Credentials() error {
	// LOAD INI CREDS
	if _, err := os.Stat(pathConfigModulesCredentials); err == nil {
		config, err := ini.Load(pathConfigModulesCredentials)
		if err != nil {
			return fmt.Errorf("failed to parse module credentials file: %s", err)
		} else {
			flickrKey = config.Section("").Key("flickr_key").String()

			instagramEmail = config.Section("").Key("instagram_email").String()
			instagramPassword = config.Section("").Key("instagram_password").String()

			spotifyClientID = config.Section("").Key("spotify_client_id").String()
			spotifyClientSecret = config.Section("").Key("spotify_client_secret").String()

			twitterAccessToken = config.Section("").Key("twitter_access_token").String()
			twitterAccessSecret = config.Section("").Key("twitter_access_secret").String()
			twitterConsumerKey = config.Section("").Key("twitter_consumer_key").String()
			twitterConsumerSecret = config.Section("").Key("twitter_consumer_secret").String()
		}
	} else {
		return fmt.Errorf("module credentials file not found: %s", err)
	}
	return nil
}

func loadConfig_Modules() []error {
	return []error{
		loadConfig_Module_RSS(),
		loadConfig_Module_Instagram(),
		loadConfig_Module_Twitter(),
	}
}

//#endregion
