package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	pathConfigModuleRSS = pathConfigModules + string(os.PathSeparator) + "rss.json"
	rssConfig           configModuleRSS
)

type configModuleRSS struct {
	WaitMins int                          `json:"waitMins,omitempty"`
	DayLimit int                          `json:"dayLimit,omitempty"` // X days = too old, ignored
	Tags     []string                     `json:"tags"`
	Feeds    []configModuleTwitterAccount `json:"feeds"`
}

type configModuleRSS_Feed struct {
	// MAIN
	URL         string   `json:"url"`
	Destination string   `json:"destination"`
	DisplayName string   `json:"displayName,omitempty"`
	WaitMins    *int     `json:"waitMins,omitempty"`
	Tags        []string `json:"tags"`
	IgnoreDate  *bool    `json:"ignoreDate,omitempty"`
	DisableInfo *bool    `json:"disableInfo,omitempty"`

	// APPEARANCE
	AvatarURL            *string `json:"avatarURL,omitempty"`
	UseTwitterAppearance *string `json:"useTwitterAppearance,omitempty"`

	// RULES
	Blacklist        []string `json:"blacklist"`
	BlacklistDomains []string `json:"blacklistDomains"`
	BlacklistURL     []string `json:"blacklistURL"`
	Whitelist        []string `json:"whitelist"`
}

func loadConfig_Module_RSS() error {

	// LOAD JSON CONFIG
	if _, err := os.Stat(pathConfigModuleRSS); err != nil {
		return fmt.Errorf("rss config file not found: %s", err)
	} else {
		configBytes, err := os.ReadFile(pathConfigModuleRSS)
		if err != nil {
			return fmt.Errorf("failed to read rss config file: %s", err)
		} else {
			// Fix backslashes
			configStr := string(configBytes)
			configStr = strings.ReplaceAll(configStr, "\\", "\\\\")
			for strings.Contains(configStr, "\\\\\\") {
				configStr = strings.ReplaceAll(configStr, "\\\\\\", "\\\\")
			}
			// Parse
			if err = json.Unmarshal([]byte(configStr), &rssConfig); err != nil {
				return fmt.Errorf("failed to parse rss config file: %s", err)
			}
			// Output?
			if generalConfig.OutputSettings {
				s, err := json.MarshalIndent(rssConfig, "", "\t")
				if err != nil {
					log.Println(color.HiRedString("failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString("loadConfig_Module_RSS():\n%s", color.YellowString(string(s))))
				}
			}
		}
	}

	return nil
}

func handleRSS_Feed(feed configModuleRSS_Feed) {
	//
}
