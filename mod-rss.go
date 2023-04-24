package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gtuk/discordwebhook"
	"github.com/mmcdole/gofeed"
)

var (
	pathConfigModuleRSS = pathConfigModules + string(os.PathSeparator) + "rss.json"
	rssConfig           configModuleRSS

	moduleNameRSS = "rss"
)

type configModuleRSS struct {
	WaitMins int                    `json:"waitMins,omitempty"`
	DayLimit int                    `json:"dayLimit,omitempty"` // X days = too old, ignored
	Tags     []string               `json:"tags,omitempty"`
	Feeds    []configModuleRSS_Feed `json:"feeds"`
}

type configModuleRSS_Feed struct {
	// MAIN
	URL          string   `json:"url"`
	Destinations []string `json:"destinations"`
	Name         string   `json:"name,omitempty"`
	WaitMins     *int     `json:"waitMins,omitempty"`
	//Tags         []string `json:"tags,omitempty"`
	//IgnoreDate   *bool    `json:"ignoreDate,omitempty"`
	//DisableInfo  *bool    `json:"disableInfo,omitempty"`

	// APPEARANCE
	Avatar     *string `json:"avatar,omitempty"`
	UseTwitter *string `json:"useTwitter,omitempty"`

	// RULES
	Blacklist    [][]string `json:"blacklist,omitempty"`
	BlacklistURL [][]string `json:"blacklistURL,omitempty"`
	Whitelist    [][]string `json:"whitelist,omitempty"`
	ListType     string     `json:"listType,omitempty"`
	//BlacklistDomains [][]string `json:"blacklistDomains,omitempty"`
}

func loadConfig_Module_RSS() error {
	prefixHere := "loadConfig_Module_RSS(): "

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
					log.Println(color.HiRedString(prefixHere+"failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString(prefixHere+"\n%s", color.YellowString(string(s))))
				}
			}
		}
	}

	return nil
}

func handleRSS_Feed(feed configModuleRSS_Feed) error {
	prefixHere := fmt.Sprintf("handleRSS_Feed(\"%s\"): ", feed.URL)
	log.Println(color.BlueString("(DEBUG) EVENT FIRED ~ RSS: %s", feed.URL))
	//
	fp := gofeed.NewParser()
	fp.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"
	rss, err := fp.ParseURL(feed.URL)
	if err != nil {
		return fmt.Errorf(prefixHere+"error parsing rss feed: %s", err.Error())
	} else {
		username := rss.Title
		if feed.Name != "" {
			username = feed.Name
		}
		avatar := ""

		if feed.UseTwitter != nil {
			if twitterClient == nil {
				return errors.New(prefixHere + "feed uses Twitter for appearance but Twitter client is empty")
			}
			id64, err := strconv.ParseInt(*feed.UseTwitter, 10, 64)
			if err != nil {
				return fmt.Errorf(prefixHere+"feed uses Twitter for appearance but error converting ID to int64: %s", err.Error())
			}
			twitterUsers, err := twitterClient.GetUsersLookupByIds([]int64{id64}, url.Values{})
			if err != nil {
				return fmt.Errorf(prefixHere+"feed uses Twitter for appearance but failed to fetch twitter user: %s", err.Error())
			}
			twitterUser := twitterUsers[0]
			username = twitterUser.Name
			avatar = strings.ReplaceAll(twitterUser.ProfileImageUrlHttps, "_normal", "_400x400")
		}

		if feed.Avatar != nil {
			avatar = *feed.Avatar
		}

		// FOREACH Entry
		for i := len(rss.Items) - 1; i >= 0; i-- { // process oldest to newest
			entry := rss.Items[i]
			link := entry.Link

			// SETUP CHECK
			vibeCheck := true
			if len(feed.Blacklist) > 0 && len(feed.Whitelist) > 0 && feed.ListType != "" {
				if feed.ListType == "wb" {
					vibeCheck = false
				} else /*if feed.ListType == "bw"*/ {
					vibeCheck = true
				}
			} else if len(feed.Blacklist) > 0 {
				vibeCheck = true
			} else if len(feed.Whitelist) > 0 {
				vibeCheck = false
			}
			checkOtherBlacklist := func(ok bool, blacklist [][]string, haystack string) bool {
				for _, row := range blacklist {
					if !ok {
						break
					}
					if containsAll(haystack, row) {
						ok = false
					}
				}
				return ok
			}
			checkBlacklist := func(ok bool, haystack string) bool {
				for _, row := range feed.Blacklist {
					if !ok {
						break
					}
					if containsAll(haystack, row) {
						ok = false
					}
				}
				return ok
			}
			checkWhitelist := func(ok bool, haystack string) bool {
				for _, row := range feed.Whitelist {
					if ok {
						break
					}
					if containsAll(haystack, row) {
						ok = true
					}
				}
				return ok
			}
			checkLists := func(ok bool, haystack string) bool {
				if feed.ListType == "wb" {
					if len(feed.Whitelist) > 0 {
						ok = checkWhitelist(ok, haystack)
					}
					if len(feed.Blacklist) > 0 {
						ok = checkBlacklist(ok, haystack)
					}
				} else /*if feed.ListType == "bw"*/ {
					if len(feed.Blacklist) > 0 {
						ok = checkBlacklist(ok, haystack)
					}
					if len(feed.Whitelist) > 0 {
						ok = checkWhitelist(ok, haystack)
					}
				}
				return ok
			}

			vibeCheck = checkLists(vibeCheck, entry.Title)
			vibeCheck = checkLists(vibeCheck, entry.Content)
			if len(feed.BlacklistURL) > 0 {
				vibeCheck = checkOtherBlacklist(vibeCheck, feed.BlacklistURL, link)
			}

			/*var colorFunc func(string, ...interface{}) string
			if vibeCheck {
				colorFunc = color.HiGreenString
			} else {
				colorFunc = color.HiRedString
			}
			log.Println(colorFunc("RSS: %s %s\n\t\t\"%s\"", entry.Updated, link, entry.Title))*/

			if vibeCheck { //TODO: AND meets days old criteria
				for _, destination := range feed.Destinations {
					if !refCheckSentToChannel(link, destination) {
						// SEND
						err = sendWebhook(destination, link, discordwebhook.Message{
							Username:  &username,
							AvatarUrl: &avatar,
							Content:   &link,
						}, moduleNameRSS)
						if err != nil {
							// we want it to process the rest, so no err return
							log.Println(color.HiRedString(prefixHere+"error sending webhook message: %s", err.Error()))
						}
					}
				}
			}
		}
	}

	return nil
}
