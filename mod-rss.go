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
	Username   *string `json:"username,omitempty"`
	Avatar     *string `json:"avatar,omitempty"`
	UseTwitter *string `json:"useTwitter,omitempty"`

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
	prefixHere := fmt.Sprintf("handleRssFeed(\"%s\"): ", feed.URL)
	log.Println(color.BlueString("(DEBUG) EVENT FIRED ~ RSS: %s", feed.URL))
	//
	fp := gofeed.NewParser()
	fp.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"
	rss, err := fp.ParseURL(feed.URL)
	if err != nil {
		return fmt.Errorf(prefixHere+"error parsing rss feed: %s", err.Error())
	} else {
		username := rss.Title
		avatar := ""

		//TODO: FIX THIS
		/*if feed.UseTwitter != nil {
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
		}*/

		if feed.Username != nil {
			username = *feed.Username
		}
		if feed.Avatar != nil {
			avatar = *feed.Avatar
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
						reply := tags + link
						// SEND
						err = sendWebhook(destination.Channel, link, discordwebhook.Message{
							Username:  &username,
							AvatarUrl: &avatar,
							Content:   &reply,
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
		val := opt.StringValue()
		config.Avatar = &val
	}
	if opt, ok := optionMap["username"]; ok {
		val := opt.StringValue()
		config.Username = &val
	}
	//TODO: FIX THIS
	/*if opt, ok := optionMap["twitter"]; ok {
		if twitterClient == nil {
			//TODO: log
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "Twitter Client is not connected..."},
			})
			return errors.New("twitter client is nil")
		} else {
			handle := opt.StringValue()
			userResults, err := twitterClient.GetUsersLookup(handle, url.Values{})
			if err == nil {
				//TODO: log
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: "ERROR FETCHING USERS..."},
				})
				return fmt.Errorf("error fetching users: %s", err.Error())
			} else {
				if len(userResults) > 0 {
					config.UseTwitter = &userResults[0].IdStr
				} else {
					//TODO: log
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "No Twitter users found for this handle..."},
					})
					return errors.New("no twitter users found for this handle")
				}
			}
		}
	}*/
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
