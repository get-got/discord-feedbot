package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"github.com/fatih/color"
)

var (
	pathConfigModuleTwitter = pathConfigModules + string(os.PathSeparator) + "twitter.json"
	twitterConfig           configModuleTwitter
)

type configModuleTwitter struct {
	WaitMins int                          `json:"waitMins,omitempty"`
	Accounts []configModuleTwitterAccount `json:"accounts"`
}

type configModuleTwitterAccount struct {
	ID          string `json:"id"`
	Destination string `json:"destination"`
	WaitMins    *int   `json:"waitMins,omitempty"`
}

func loadConfig_Module_Twitter() error {
	// TODO: Creation prompts if missing

	// LOAD JSON CONFIG
	if _, err := os.Stat(pathConfigModuleTwitter); err != nil {
		return fmt.Errorf("twitter config file not found: %s", err)
	} else {
		configBytes, err := os.ReadFile(pathConfigModuleTwitter)
		if err != nil {
			return fmt.Errorf("failed to read twitter config file: %s", err)
		} else {
			// Fix backslashes
			configStr := string(configBytes)
			configStr = strings.ReplaceAll(configStr, "\\", "\\\\")
			for strings.Contains(configStr, "\\\\\\") {
				configStr = strings.ReplaceAll(configStr, "\\\\\\", "\\\\")
			}
			// Parse
			if err = json.Unmarshal([]byte(configStr), &twitterConfig); err != nil {
				return fmt.Errorf("failed to parse twitter config file: %s", err)
			}
			// Output?
			if generalConfig.OutputSettings {
				s, err := json.MarshalIndent(twitterConfig, "", "\t")
				if err != nil {
					log.Println(color.HiRedString("failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString("loadConfig_Module_Twitter():\n%s", color.YellowString(string(s))))
				}
			}
		}
	}

	return nil
}

var (
	twitterAccessToken    string
	twitterAccessSecret   string
	twitterConsumerKey    string
	twitterConsumerSecret string

	twitterClient    *anaconda.TwitterApi
	twitterConnected bool = false
)

func openTwitter() error {
	if twitterAccessToken == "" || twitterAccessSecret == "" ||
		twitterConsumerKey == "" || twitterConsumerSecret == "" {
		return errors.New("twitter credentials are incomplete")
	}
	twitterClient = anaconda.NewTwitterApiWithCredentials(
		twitterAccessToken, twitterAccessSecret,
		twitterConsumerKey, twitterConsumerSecret,
	)

	twitterSelf, err := twitterClient.GetSelf(url.Values{})
	if err != nil {
		return fmt.Errorf("twitter api failed to fetch data on self: %s", err.Error())
	} else {
		twitterConnected = true
		log.Println(color.HiMagentaString("Twitter API connected to @%s", twitterSelf.ScreenName))
	}

	return nil
}

func handleTwitterAccount(account configModuleTwitterAccount) {
	log.Printf("twitter account event fired: %s", account.ID)
}
