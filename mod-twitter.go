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

	"github.com/ChimeraCoder/anaconda"
	"github.com/fatih/color"
	"github.com/gtuk/discordwebhook"
)

var (
	pathConfigModuleTwitter = pathConfigModules + string(os.PathSeparator) + "twitter.json"
	twitterConfig           configModuleTwitter

	moduleNameTwitterAccounts = "twitter-accounts"

	twitterLogo = "https://i.imgur.com/BEZiTLN.png"
)

type configModuleTwitter struct {
	Tags []string `json:"tags,omitempty"`

	WaitMins int `json:"waitMins,omitempty"`
	DayLimit int `json:"dayLimit,omitempty"` // X days = too old, ignored

	Accounts []configModuleTwitterAccount `json:"accounts"`
}

type configModuleTwitterAccount struct {
	// MAIN
	ID           string   `json:"id"`
	Destinations []string `json:"destinations"`
	Tags         []string `json:"tags,omitempty"`

	WaitMins *int `json:"waitMins,omitempty"`
	DayLimit *int `json:"dayLimit,omitempty"` // X days = too old, ignored

	// APPEARANCE
	MaskUsername *string `json:"maskUsername,omitempty"`
	MaskAvatar   *string `json:"maskAvatar,omitempty"`
	MaskColor    *string `json:"maskColor,omitempty"`

	// RULES
	ExcludeReplies    *bool      `json:"excludeReplies,omitempty"`
	IncludeRetweets   *bool      `json:"includeRetweets,omitempty"`
	FilterType        string     `json:"filterType,omitempty"`
	ListType          string     `json:"listType,omitempty"`
	Blacklist         [][]string `json:"blacklist"`
	Whitelist         [][]string `json:"whitelist"`
	BlacklistRetweets []string   `json:"blacklistRetweetsFrom"`
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

func handleTwitterAccount(account configModuleTwitterAccount) error {
	log.Println(color.BlueString("(DEBUG) EVENT FIRED ~ TWITTER ACCOUNT: %s", account.ID))

	if twitterClient == nil {
		return errors.New("twitter client is invalid")
	}

	// Vars
	includeRetweets := false
	if account.IncludeRetweets != nil {
		includeRetweets = *account.IncludeRetweets
	}
	excludeReplies := false
	if account.ExcludeReplies != nil {
		excludeReplies = *account.ExcludeReplies
	}

	// User Info
	id64, err := strconv.ParseInt(account.ID, 10, 64)
	if err != nil {
		return fmt.Errorf("error converting ID to int64: %s", err.Error())
	}
	userInfo, err := twitterClient.GetUsersLookupByIds([]int64{id64}, url.Values{})
	if err != nil {
		return fmt.Errorf("[ID:%s] failed to fetch twitter user: %s", account.ID, err.Error())
	}
	if len(userInfo) == 0 {
		return fmt.Errorf("[ID:%s] no users found", account.ID)
	}
	user := userInfo[0]

	// User Appearance Vars
	handle := user.ScreenName
	username := user.Name
	if account.MaskUsername != nil {
		username = *account.MaskUsername
	}
	avatar := strings.ReplaceAll(user.ProfileImageUrlHttps, "_normal", "_400x400")
	if account.MaskAvatar != nil {
		avatar = *account.MaskAvatar
	}
	userColor := user.ProfileLinkColor
	if account.MaskColor != nil {
		userColor = *account.MaskColor
	}

	// User Timeline
	tmpArgs := url.Values{}
	tmpArgs.Add("user_id", account.ID)
	tmpArgs.Add("count", "35")
	tmpArgs.Add("include_rts", strconv.FormatBool(includeRetweets))
	tmpArgs.Add("exclude_replies", "false")
	tmpArgs.Add("tweet_mode", "extended")
	tweets, err := twitterClient.GetUserTimeline(tmpArgs)
	if err != nil {
		return fmt.Errorf("[ID:%s] error fetching timeline: %s", account.ID, err.Error())
	}
	if len(tweets) == 0 {
		return fmt.Errorf("[ID:%s] timeline has no tweets", account.ID)
	}

	// FOREACH Tweet
	for i := len(tweets) - 1; i >= 0; i-- { // process oldest to newest
		// Tweet Vars
		//TODO: calc & check timespan
		tweet := tweets[i]
		//tweetPathS := handle + "/" + tweet.IdStr
		tweetPath := handle + "/status/" + tweet.IdStr
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

		vibeCheck = checkLists(vibeCheck, tweet.FullText)
		//TODO: check media titles
		// THREAD CHECKS
		if excludeReplies && tweet.FullText[:1] == "@" {
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
				if strings.Contains(tweet.FullText, fmt.Sprintf("RT %s: ", handle)) {
					vibeCheck = false
					break
				}
			}
		}

		// Embed Vars
		embedDesc := tweet.FullText
		embedFooterText := "Xm ago, Y likes, Z retweets"
		embedColor := hexdec(userColor)
		log.Println(embedColor)

		//TODO: Output
		/*var colorFunc func(string, ...interface{}) string
		if vibeCheck {
			colorFunc = color.HiGreenString
		} else {
			colorFunc = color.HiRedString
		}
		log.Println(colorFunc("TWEET: %s %s\n\t\t\"%s\"", tweet.CreatedAt, tweetPathS, tweet.FullText))*/

		// PROCESS
		if vibeCheck { //TODO: AND meets days old criteria
			for _, destination := range account.Destinations {
				if !refCheckSentToChannel(tweetLink, destination) {
					// SEND
					sendWebhook(destination, tweetLink, discordwebhook.Message{
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
				}
			}
		}
	}

	return nil
}
