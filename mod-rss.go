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
	"github.com/gtuk/discordwebhook"
	"github.com/mmcdole/gofeed"
)

var (
	pathConfigModuleRSS = pathConfigModules + string(os.PathSeparator) + "rss.json"
	rssConfig           configModuleRSS

	moduleNameRSS = "rss"
)

type configModuleRSS struct {
	WaitMins int `json:"waitMins,omitempty"`
	DayLimit int `json:"dayLimit,omitempty"` // X days = too old, ignored

	Feeds []configModuleRssFeed `json:"feeds"`
}

type configModuleRssFeed struct {
	// MAIN
	Name         string            `json:"name"`
	URL          string            `json:"url"`
	Destinations []feedDestination `json:"destinations"`

	WaitMins *int `json:"waitMins,omitempty"`
	//IgnoreDate   *bool    `json:"ignoreDate,omitempty"`
	//DisableInfo  *bool    `json:"disableInfo,omitempty"`

	// APPEARANCE
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Twitter  string `json:"twitter,omitempty"`

	// GENERIC RULES
	Blacklist [][]string `json:"blacklist,omitempty"`
	Whitelist [][]string `json:"whitelist,omitempty"`
	ListType  string     `json:"listType,omitempty"`
	// + LIST RULES
	BlacklistURL [][]string `json:"blacklistURL,omitempty"`
	//BlacklistDomains [][]string `json:"blacklistDomains,omitempty"`
	// RULES
	//.
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

func handleRssFeed(feed configModuleRssFeed) error {
	l := logInstructions{
		Location: fmt.Sprintf("handleRssFeed(@%s): ", feed.Name),
		Task:     "",
		Inline:   false,
		Color:    color.BlueString,
	}
	if generalConfig.Debug {
		log.Println(l.SetFlag(&lDebug).LogI(true, "FEED STARTING ... RSS Feed \"%s\"", feed.Name))
		l.ClearFlag()
	}
	//
	fp := gofeed.NewParser()
	fp.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"
	rss, err := fp.ParseURL(feed.URL)
	if err != nil {
		return fmt.Errorf(feed.Name+": error parsing rss feed: %s", err.Error())
	} else {
		username := rss.Title
		avatar := ""

		if feed.Twitter != "" {
			handle := feed.Twitter
			if cachedAvatar, exists := twitterAvatarCache[handle]; exists {
				avatar = cachedAvatar
			} else {
				twitterUser, err := twitterScraper.GetProfile(handle)
				if err != nil {
					return fmt.Errorf(feed.Name+": feed uses Twitter for appearance but failed to fetch twitter user: %s", err.Error())
				}
				username = twitterUser.Name
				avatar = strings.ReplaceAll(twitterUser.Avatar, "_normal", "_400x400")
				twitterAvatarCache[handle] = avatar
			}
		}

		// User Appearance Vars
		if feed.Username != "" {
			username = feed.Username
		}
		if feed.Avatar != "" {
			avatar = feed.Avatar
		}

		if generalConfig.Debug2 {
			log.Println(l.SetFlag(&lDebug2).LogI(true, "FEED PARSED ... %d items - titled \"%s\"", len(rss.Items), rss.Title))
			l.ClearFlag()
		}

		// FOREACH Entry
		for i := len(rss.Items) - 1; i >= 0; i-- { // process oldest to newest
			entry := rss.Items[i]
			link := entry.Link
			// Unwrap Google Links
			if strings.Contains(link, "&url=") && strings.Contains(link, "&ct=") {
				link = link[strings.Index(link, "&url=")+5:]
				link = link[:strings.Index(link, "&ct=")]
			}

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
					sendAttempts := 0
					if !refCheckSentToChannel(link, destination.Channel) {
						tags := ""
						for _, tag := range destination.Tags {
							if tags == "" {
								tags = fmt.Sprintf("<@%s>", tag)
							} else {
								tags += fmt.Sprintf(", <@%s>", tag)
							}
						}
						if tags != "" {
							tags += "\n"
						}
						//TODO: Published X \n link
						reply := tags + link
						// SEND
					resend:
						sendAttempts++
						webhookInfo := fmt.Sprintf("WEBHOOK to %s (\"%s\")", destination.Channel, link)
						// SEND
						err = sendWebhook(destination.Channel, link, discordwebhook.Message{
							Username:  &username,
							AvatarUrl: &avatar,
							Content:   &reply,
						}, moduleNameRSS)
						if err != nil {
							// we want it to process the rest, so no err return
							//TODO: implement this universally vvvvvvvv
							if strings.Contains(err.Error(), "resource is being rate limited") {
								log.Println(l.SetFlag(&lError).Log(
									"%s is being rate limited... delaying 3 seconds and trying again...", webhookInfo))
								l.ClearFlag()
								time.Sleep(3 * time.Second)
								if sendAttempts < 5 {
									goto resend
								} else {
									log.Println(l.SetFlag(&lError).Log(
										"%s was rate limited more than 5 times, giving up...", webhookInfo))
									l.ClearFlag()
								}
								//TODO: ^^^^^^^
							} else {
								log.Println(l.SetFlag(&lError).Log(
									"%s encountered an error while sending: %s", webhookInfo, err.Error()))
								l.ClearFlag()
							}
						} else if generalConfig.Debug2 {
							log.Println(l.SetFlag(&lDebug2).LogI(true, "SENT %s to %s", link, destination.Channel))
							l.ClearFlag()
						}
					} else if generalConfig.Debug2 {
						log.Println(l.SetFlag(&lDebug2).LogCI(color.BlueString, true, "- ALREADY SENT %s to %s", link, destination.Channel))
						l.ClearFlag()
					}
				}
			}
		}
	}

	if generalConfig.Debug {
		waitMins := rssConfig.WaitMins
		if feed.WaitMins != nil {
			waitMins = *feed.WaitMins
		}
		log.Println(l.SetFlag(&lDebug).LogI(true, "FEED COMPLETED ... RSS Feed %s ... waiting %d minutes", feed.Name, waitMins))
		l.ClearFlag()
	}

	return nil
}

func handleRssCmdOpts(config *configModuleRssFeed,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
	s *discordgo.Session, i *discordgo.InteractionCreate) error {

	// Optional Vars
	if opt, ok := optionMap["change-url"]; ok {
		config.URL = opt.StringValue()
	}
	if opt, ok := optionMap["tag"]; ok {
		tagged := opt.UserValue(s)
		if tagged != nil {
			destClone := config.Destinations
			for key, destination := range destClone {
				if destination.Channel == i.ChannelID {
					config.Destinations[key].Tags = []string{tagged.ID}
				}
			}
		}
	}
	if opt, ok := optionMap["wait"]; ok {
		val := int(opt.IntValue())
		config.WaitMins = &val
	}
	if opt, ok := optionMap["avatar"]; ok {
		config.Avatar = opt.StringValue()
	}
	if opt, ok := optionMap["username"]; ok {
		config.Username = opt.StringValue()
	}
	if opt, ok := optionMap["twitter"]; ok {
		config.Twitter = opt.StringValue()
	}
	// Optional Vars -Lists
	if opt, ok := optionMap["blacklist"]; ok {
		var list []string
		list = append(list, strings.Split(opt.StringValue(), "|")...)
		config.Blacklist = append(config.Blacklist, list)
	}
	if opt, ok := optionMap["whitelist"]; ok {
		var list []string
		list = append(list, strings.Split(opt.StringValue(), "|")...)
		config.Whitelist = append(config.Whitelist, list)
	}
	if opt, ok := optionMap["list-type"]; ok {
		config.ListType = opt.StringValue()
	}
	if opt, ok := optionMap["blacklist-url"]; ok {
		var list []string
		list = append(list, strings.Split(opt.StringValue(), "|")...)
		config.BlacklistURL = append(config.BlacklistURL, list)
	}
	return nil
}

func getRssConfigIndex(name string) int {
	for k, feed := range rssConfig.Feeds {
		if strings.EqualFold(name, feed.Name) {
			return k
		}
	}
	return -1
}

func getRssConfig(name string) *configModuleRssFeed {
	i := getRssConfigIndex(name)
	if i == -1 {
		return nil
	} else {
		return &rssConfig.Feeds[i]
	}
}

func existsRssConfig(name string) bool {
	return getRssConfig(name) != nil
}

func updateRssConfig(name string, config configModuleRssFeed) bool {
	feedClone := rssConfig.Feeds
	for key, feed := range feedClone {
		if strings.EqualFold(name, feed.Name) {
			rssConfig.Feeds[key] = config
			return true
		}
	}
	return false
}

func deleteRssConfig(name string) error {
	index := getRssConfigIndex(name)
	if index != -1 {
		// Remove from loaded config
		rssConfig.Feeds = append(rssConfig.Feeds[:index], rssConfig.Feeds[index+1:]...)
		// Remove from live feeds
		if !deleteFeed(name, feedRSS) {
			return errors.New("failed to delete from live feeds")
		}
		return nil
	}
	return errors.New("rss config does not exist")
}
