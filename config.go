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

func loadConfig() map[string]error {
	errors := make(map[string]error)

	var tmperr error

	if tmperr = loadConfig_General(); tmperr != nil {
		errors["general"] = tmperr
	}

	if tmperr = loadConfig_Discord(); tmperr != nil {
		errors["discord"] = tmperr
	}

	if tmperr = loadConfig_Modules_Credentials(); tmperr != nil {
		errors["mod-credentials"] = tmperr
	}

	tmperrs := loadConfig_Modules()
	for module, err := range tmperrs {
		if err != nil {
			errors[module] = err
		}
	}

	return errors
}

func saveConfig(filepath string, i interface{}) error {
	json, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath, json, 0644)
	if err != nil {
		return err
	}
	return nil
}

//#endregion

//#region General

var (
	pathConfigGeneralSettings    = pathConfig + string(os.PathSeparator) + "general.json"
	pathConfigDiscordCredentials = pathConfig + string(os.PathSeparator) + "discord.ini"
	pathConfigDiscordSettings    = pathConfig + string(os.PathSeparator) + "discord.json"
)

func loadConfig_General() error {
	prefixHere := "loadConfig_General(): "
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

			twitterUsername = config.Section("").Key("twitter_username").String()
			twitterPassword = config.Section("").Key("twitter_password").String()
		}
	} else {
		return fmt.Errorf("module credentials file not found: %s", err)
	}
	return nil
}

func loadConfig_Modules() map[string]error {
	return map[string]error{
		"mod-rss":       loadConfig_Module_RSS(),
		"mod-instagram": loadConfig_Module_Instagram(),
		"mod-twitter":   loadConfig_Module_Twitter(),
	}
}

//#endregion
