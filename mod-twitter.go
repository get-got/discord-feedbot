package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/gtuk/discordwebhook"
	twitterscraper "github.com/n0madic/twitter-scraper"
)

var (
	pathConfigModuleTwitter = pathConfigModules + string(os.PathSeparator) + "twitter.json"
	twitterConfig           configModuleTwitter

	moduleNameTwitterAccounts = "twitter-accounts"

	twitterLogo = "https://i.imgur.com/BEZiTLN.png"

	twitterAvatarCache = make(map[string]string)
)

type configModuleTwitter struct {
	OverwriteCache string `json:"overwriteCache,omitempty"`

	WaitMins int `json:"waitMins,omitempty"`
	//DayLimit int `json:"dayLimit,omitempty"` // X days = too old, ignored

	DefaultColor string `json:"defaultColor,omitempty"`

	Accounts []configModuleTwitterAcc `json:"accounts"`
}

type configModuleTwitterAcc struct {
	// MAIN
	Name         string            `json:"name"`
	Handle       string            `json:"handle"`
	Destinations []feedDestination `json:"destinations"`

	WaitMins *int `json:"waitMins,omitempty"`
	//DayLimit *int `json:"dayLimit,omitempty"` // X days = too old, ignored

	// APPEARANCE
	Username string `json:"username,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Color    string `json:"color,omitempty"`

	// GENERIC RULES
	Blacklist [][]string `json:"blacklist"`
	Whitelist [][]string `json:"whitelist"`
	ListType  string     `json:"listType,omitempty"`
	// + LIST RULES
	BlacklistRetweets []string `json:"blacklistRetweetsFrom"` //TODO: command control
	// RULES
	ExcludeReplies  *bool  `json:"excludeReplies,omitempty"`
	IncludeRetweets *bool  `json:"includeRetweets,omitempty"`
	FilterType      string `json:"filterType,omitempty"`
}

func loadConfig_Module_Twitter() error {
	prefixHere := "loadConfig_Module_Twitter(): "
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
					log.Println(color.HiRedString(prefixHere+"failed to output...\t%s", err))
				} else {
					log.Println(color.HiYellowString(prefixHere+"\n%s", color.YellowString(string(s))))
				}
			}

			// Overwrite Cookies?
			if twitterConfig.OverwriteCache != "" {
				pathDataCookiesTwitter = twitterConfig.OverwriteCache
			}
		}
	}

	return nil
}

var (
	twitterUsername  string
	twitterPassword  string
	twitterConnected bool = false

	twitterScraper *twitterscraper.Scraper
)

func openTwitter() error {

	twitterImport := func() error {
		f, err := os.Open(pathDataCookiesTwitter)
		if err != nil {
			return err
		}
		var cookies []*http.Cookie
		err = json.NewDecoder(f).Decode(&cookies)
		if err != nil {
			return err
		}
		twitterScraper.SetCookies(cookies)
		twitterScraper.IsLoggedIn()
		_, err = twitterScraper.GetProfile("x")
		if err != nil {
			return err
		}
		return nil
	}

	twitterExport := func() error {
		cookies := twitterScraper.GetCookies()
		js, err := json.Marshal(cookies)
		if err != nil {
			return err
		}
		f, err := os.Create(pathDataCookiesTwitter)
		if err != nil {
			return err
		}
		f.Write(js)
		return nil
	}

	twitterScraper = twitterscraper.New()

	if twitterUsername != "" &&
		twitterPassword != "" {
		log.Println(color.MagentaString("Connecting to Twitter (X)..."))

		twitterLoginCount := 0
	do_twitter_login:
		twitterLoginCount++
		if twitterLoginCount > 1 {
			time.Sleep(3 * time.Second)
		}

		if twitterImport() != nil {
			twitterScraper.ClearCookies()
			if err := twitterScraper.Login(twitterUsername, twitterPassword); err != nil {
				log.Println(color.HiRedString("Login Error: %s", err.Error()))
				if twitterLoginCount <= 3 {
					goto do_twitter_login
				} else {
					log.Println(color.HiRedString(
						"Failed to login to Twitter (X), the bot will not fetch this media..."))
				}
			} else {
				twitterConnected = true
				defer twitterExport()
				log.Println(color.HiMagentaString("Connected"))
				if twitterScraper.IsLoggedIn() {
					log.Println(color.HiMagentaString("Connected to @%s via new login", twitterUsername))
				} else {
					log.Println(color.HiRedString("Scraper login seemed successful but bot is not logged in, Twitter (X) parsing may not work..."))
				}
			}
		} else {
			log.Println(color.HiMagentaString("Connected to @%s via cache", twitterUsername))
			twitterConnected = true
		}
	} else {
		log.Println(color.MagentaString("Twitter (X) credentials missing, the bot will not fetch this media..."))
	}

	return nil
}

func handleTwitterAcc(account configModuleTwitterAcc) error {
	prefixHere := fmt.Sprintf("handleTwitterAccount(%s): ", account.Handle)
	log.Println(color.BlueString("(DEBUG) EVENT FIRED ~ TWITTER ACCOUNT: %s @%s", account.Name, account.Handle))

	// Vars
	/*includeRetweets := false
	if account.IncludeRetweets != nil {
		includeRetweets = *account.IncludeRetweets
	}*/
	excludeReplies := false
	if account.ExcludeReplies != nil {
		excludeReplies = *account.ExcludeReplies
	}

	// User Info
	user, err := twitterScraper.GetProfile(account.Handle)
	if err != nil {
		return fmt.Errorf("[ID:%s] failed to fetch twitter user @%s", account.Handle, err.Error())
	}

	// User Appearance Vars
	handle := account.Handle
	username := user.Name
	if account.Username != "" {
		username = account.Username
	}
	avatar := strings.ReplaceAll(user.Avatar, "_normal", "_400x400")
	if account.Avatar != "" {
		avatar = account.Avatar
	}
	userColor := projectColor             // default to project
	if generalConfig.DefaultColor != "" { // override with general if present
		userColor = generalConfig.DefaultColor
	}
	if twitterConfig.DefaultColor != "" { // override with twitter if present
		userColor = twitterConfig.DefaultColor
	}
	if account.Color != "" { // override with specific if present
		userColor = account.Color
	}

	// User Timeline
	tweets := twitterScraper.GetTweets(context.Background(), account.Handle, 50)

	// FOREACH Tweet
	//for i := len(tweets) - 1; i >= 0; i-- { // process oldest to newest
	for tweet := range tweets {
		// Tweet Vars
		//TODO: calc & check timespan
		//tweetPathS := handle + "/" + tweet.IdStr
		tweetPath := handle + "/status/" + tweet.ID
		tweetLink := "https://twitter.com/" + tweetPath
		/*tweetParent := tweet
		if tweet.RetweetedStatus != nil {
			if tweet.RetweetedStatus.QuotedStatus != nil { // RT'd Quote
				tweetParent = *tweet.RetweetedStatus.QuotedStatus
			} else { // RT
				tweetParent = *tweet.RetweetedStatus
			}
		} else if tweet.QuotedStatus != nil { // Quote
			tweetParent = *tweet.QuotedStatus
		}*/
		/*var tweetMediaSource []anaconda.EntityMedia = nil
		if len(tweetParent.ExtendedEntities.Media) > 0 {
			tweetMediaSource = tweetParent.ExtendedEntities.Media
		}*/

		// SETUP CHECK
		vibeCheck := true
		if len(account.Blacklist) > 0 && len(account.Whitelist) > 0 && account.ListType != "" {
			if account.ListType == "wb" {
				vibeCheck = false
			} else /*if account.ListType == "bw"*/ {
				vibeCheck = true
			}
		} else if len(account.Blacklist) > 0 {
			vibeCheck = true
		} else if len(account.Whitelist) > 0 {
			vibeCheck = false
		}

		checkBlacklist := func(ok bool, haystack string) bool {
			for _, row := range account.Blacklist {
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
			for _, row := range account.Whitelist {
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
			if account.ListType == "wb" {
				if len(account.Whitelist) > 0 {
					ok = checkWhitelist(ok, haystack)
				}
				if len(account.Blacklist) > 0 {
					ok = checkBlacklist(ok, haystack)
				}
			} else /*if account.ListType == "bw"*/ {
				if len(account.Blacklist) > 0 {
					ok = checkBlacklist(ok, haystack)
				}
				if len(account.Whitelist) > 0 {
					ok = checkWhitelist(ok, haystack)
				}
			}
			return ok
		}

		// Init Check
		vibeCheck = checkLists(vibeCheck, tweet.Tweet.Text)

		//TODO: check media titles
		// THREAD CHECKS
		if excludeReplies && tweet.Tweet.Text[:1] == "@" {
			vibeCheck = false
		}
		//TODO: Fix/Finish Thread Checking below
		/*if vibeCheck && tweet.InReplyToStatusIdStr != "" {
			tmpArgs := url.Values{}
			tmpArgs.Add("tweet_mode", "extended")
			reply, err := twitterClient.GetTweet(tweet.InReplyToStatusID, tmpArgs)
			threadCount := 0
			if err == nil { // i think this whole thing was poorly translated from old php, revisit later
				for {
					if threadCount >= 5 || !vibeCheck {
						break
					}


					tmpArgs := url.Values{}
					tmpArgs.Add("tweet_mode", "extended")
					reply, err = twitterClient.GetTweet(reply.InReplyToStatusID, tmpArgs)
					if err != nil || reply.InReplyToStatusID == "" {
						break
					}
					threadCount++
					time.Sleep(1 * time.Second)
				}
			}
		}*/

		//TODO: Check media type
		if account.FilterType != "" {
			if account.FilterType == "media" {

			} else if account.FilterType == "text" {

			} else if account.FilterType == "images" {

			} else if account.FilterType == "videos" {

			}
		}

		// Retweet Filter TODO:check
		if len(account.BlacklistRetweets) > 0 && tweet.RetweetedStatus != nil {
			for _, handle := range account.BlacklistRetweets {
				if strings.Contains(tweet.Tweet.Text, fmt.Sprintf("RT %s: ", handle)) {
					vibeCheck = false
					break
				}
			}
		}

		// Tweet Info
		prefixLikes := "No"
		if tweet.Likes > 0 {
			prefixLikes = humanize.Comma(int64(tweet.Likes))
		}
		suffixLikes := ""
		if tweet.Likes != 1 {
			suffixLikes = "s"
		}
		prefixRetweets := "No"
		if tweet.Retweets > 0 {
			prefixRetweets = humanize.Comma(int64(tweet.Retweets))
		}
		suffixRetweets := ""
		if tweet.Retweets != 1 {
			suffixRetweets = "s"
		}
		creationTime := tweet.TimeParsed

		// Embed Vars
		embedDesc := tweet.Tweet.Text
		embedFooterText := fmt.Sprintf("%s - %s like%s, %s retweet%s",
			humanize.Time(creationTime),
			prefixLikes, suffixLikes, prefixRetweets, suffixRetweets)
		embedColor, err := hexdec(userColor)
		if err != nil {
			log.Println("Error parsing color: " + err.Error())
		}

		//TODO: Embed Author if RT

		//TODO: Embed Media

		//TODO: Output
		/*var colorFunc func(string, ...interface{}) string
		if vibeCheck {
			colorFunc = color.HiGreenString
		} else {
			colorFunc = color.HiRedString
		}
		log.Println(colorFunc("TWEET: %s %s\n\t\t\"%s\"", tweet.CreatedAt, tweetPathS, tweet.FullText))*/

		//TODO: Log (aside from message sending log)

		//TODO: Send video links after embed (poll highest bitrate)

		sendAttempts := 0
		// PROCESS
		if vibeCheck { //TODO: AND meets days old criteria
			for _, destination := range account.Destinations {
				if !refCheckSentToChannel(tweetLink, destination.Channel) {
					// SEND
				resend:
					sendAttempts++
					err = sendWebhook(destination.Channel, tweetLink, discordwebhook.Message{
						Username:  &username,
						AvatarUrl: &avatar,
						Content:   &tweetLink,
						Embeds: &[]discordwebhook.Embed{{
							Description: &embedDesc,
							Color:       &embedColor,
							Footer: &discordwebhook.Footer{
								Text:    &embedFooterText,
								IconUrl: &twitterLogo,
							},
						}},
					}, moduleNameTwitterAccounts)
					if err != nil {
						// we want it to process the rest, so no err return
						//TODO: implement this universally vvvvvvvv
						if strings.Contains(err.Error(), "resource is being rate limited") {
							log.Println(color.HiRedString(prefixHere + "webhook is being rate limited, delaying 2 seconds and trying again..."))
							time.Sleep(2 * time.Second)
							if sendAttempts < 5 {
								goto resend
							}
							//TODO: ^^^^^^^
						} else {
							log.Println(color.HiRedString(prefixHere+"error sending webhook message: %s", err.Error()))
						}
					}
				}
			}
		}
	}

	return nil
}

func handleTwitterAccCmdOpts(config *configModuleTwitterAcc,
	optionMap map[string]*discordgo.ApplicationCommandInteractionDataOption,
	s *discordgo.Session, i *discordgo.InteractionCreate) error {

	// Optional Vars
	if opt, ok := optionMap["change-handle"]; ok {
		config.Handle = opt.StringValue()
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
	// Optional Vars - Appearance
	if opt, ok := optionMap["username"]; ok {
		config.Username = opt.StringValue()
	}
	if opt, ok := optionMap["avatar"]; ok {
		config.Avatar = opt.StringValue()
	}
	if opt, ok := optionMap["color"]; ok { //TODO: conversion?
		config.Color = opt.StringValue()
	}
	// Optional Vars - Rules
	if opt, ok := optionMap["exclude-replies"]; ok {
		val := opt.BoolValue()
		config.ExcludeReplies = &val
	}
	if opt, ok := optionMap["include-retweets"]; ok {
		val := opt.BoolValue()
		config.IncludeRetweets = &val
	}
	if opt, ok := optionMap["filter-type"]; ok {
		config.FilterType = opt.StringValue()
	}
	// Optional Vars - Lists
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
	return nil
}

func getTwitterAccConfigIndex(name string) int {
	for k, feed := range twitterConfig.Accounts {
		if strings.EqualFold(name, feed.Name) {
			return k
		}
	}
	return -1
}

func getTwitterAccConfig(name string) *configModuleTwitterAcc {
	i := getTwitterAccConfigIndex(name)
	if i == -1 {
		return nil
	} else {
		return &twitterConfig.Accounts[i]
	}
}

func existsTwitterAccConfig(name string) bool {
	return getTwitterAccConfig(name) != nil
}

func updateTwitterAccConfig(name string, config configModuleTwitterAcc) bool {
	feedClone := twitterConfig.Accounts
	for key, feed := range feedClone {
		if strings.EqualFold(name, feed.Name) {
			twitterConfig.Accounts[key] = config
			return true
		}
	}
	return false
}

func deleteTwitterAccConfig(name string) error {
	index := getTwitterAccConfigIndex(name)
	if index != -1 {
		// Remove from loaded config
		twitterConfig.Accounts = append(twitterConfig.Accounts[:index], twitterConfig.Accounts[index+1:]...)
		// Remove from live feeds
		if !deleteFeed(name, feedTwitterAccount) {
			return errors.New("failed to delete from live feeds")
		}
		return nil
	}
	return errors.New("twitter account config does not exist")
}
