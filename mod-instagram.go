package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	pathConfigModuleInstagram = pathConfigModules + string(os.PathSeparator) + "instagram.json"
	instagramConfig           configModuleInstagram
)

type configModuleInstagram struct {
	WaitMins int                            `json:"waitMins,omitempty"`
	Accounts []configModuleInstagramAccount `json:"accounts"`
}

type configModuleInstagramAccount struct {
	ID          string `json:"id"`
	Destination string `json:"destination"`
	WaitMins    *int   `json:"waitMins,omitempty"`
}

func loadConfig_Module_Instagram() error {
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
					log.Println(color.HiRedString("failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString("loadConfig_Module_Instagram():\n%s", color.YellowString(string(s))))
				}
			}
		}
	}

	return nil
}

var (
	instagramEmail    string
	instagramPassword string
)

func openInstagram() error {
	if instagramEmail == "" || instagramPassword == "" {
		return errors.New("instagram credentials are incomplete")
	}
	//TODO: ig login requires cache directories
	return nil
}

func handleInstagramAccount(account configModuleInstagramAccount) {
	log.Printf("instagram account event fired: %s", account.ID)
}
