package main

/*

* Twitter Tweets
* RSS
* Instagram
* Spotify Artist Releases
* Flickr

*M System Monitor

*L Twitter Trends
*L NASA APOD
*L Plex Titles
*L Twitch Chat Track
*L Twitch Live
*L Spotify Playlist Changes

 */

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
)

var (
	// General
	loop         chan os.Signal
	timeLaunched time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	timeLaunched = time.Now()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))

	//TODO: Github Update Check

	// Load Configs
	settingsErrors := loadConfig()
	if len(settingsErrors) > 0 {
		log.Println(color.HiRedString("Detected errors in settings..."))
		for _, err := range settingsErrors {
			if err != nil {
				log.Println(color.HiRedString("ERROR: %s", err))
			}
		}
	}

	//TODO: Database Load, create if missing
	loadDatabase()
}

func main() {

	if err := openDiscord(); err != nil {
		log.Println(color.HiRedString("Discord - Login Error: %s", err))
	}
	for _, err := range openAPIs() {
		if err != nil {
			log.Println(color.HiRedString("API - Login Error: %s", err))
		}
	}

	startFeeds()
	// This will need to be heavily modified to allow for live changes to configs
	feedsCopy := feeds
	for key, feed := range feedsCopy {
		go func(key int, feed moduleFeed) {
			for {
				feeds[key].timesRan++
				feeds[key].lastRan = time.Now()
				switch feed.moduleType {
				case feedRSS_Feed:
					{
						go handleRSS_Feed(feed.moduleConfig.(configModuleRSS_Feed))
					}
				case feedInstagramAccount:
					{
						go handleInstagramAccount(feed.moduleConfig.(configModuleInstagramAccount))
					}
				case feedTwitterAccount:
					{
						go handleTwitterAccount(feed.moduleConfig.(configModuleTwitterAccount))
					}
				}
				time.Sleep(feed.waitMins)
			}
		}(key, feed)
	}

	if generalConfig.Debug {
		log.Println(color.HiCyanString("Startup finished, took %s...", uptime()))
	}

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	log.Println(color.HiRedString("Exiting... "))
}

func openAPIs() []error {
	var errors []error
	var tmperr error

	if tmperr = openInstagram(); tmperr != nil {
		errors = append(errors, tmperr)
	}

	if tmperr = openTwitter(); tmperr != nil {
		errors = append(errors, tmperr)
	}

	return errors
}

const (
	feed000 = iota

	feedTwitterAccount
	feedInstagramAccount
	feedRSS_Feed
)

var feeds []moduleFeed

type moduleFeed struct { // i.e. thread, account, source, etc. sub of module
	moduleType   int
	moduleConfig interface{} // point to parent
	waitMins     time.Duration
	lastRan      time.Time
	timesRan     int
}

func startFeeds() {
	// RSS Feeds
	for _, feed := range rssConfig.Feeds {
		waitMins := time.Duration(rssConfig.WaitMins)
		if feed.WaitMins != nil {
			waitMins = time.Duration(*feed.WaitMins)
		}

		feeds = append(feeds, moduleFeed{
			moduleType:   feedRSS_Feed,
			moduleConfig: feed,
			waitMins:     waitMins * time.Minute,
		})
	}
	// Instagram, Accounts
	for _, account := range instagramConfig.Accounts {
		waitMins := time.Duration(instagramConfig.WaitMins)
		if account.WaitMins != nil {
			waitMins = time.Duration(*account.WaitMins)
		}

		feeds = append(feeds, moduleFeed{
			moduleType:   feedInstagramAccount,
			moduleConfig: account,
			waitMins:     waitMins * time.Minute,
		})
	}
	// Twitter, Accounts
	for _, account := range twitterConfig.Accounts {
		waitMins := time.Duration(twitterConfig.WaitMins)
		if account.WaitMins != nil {
			waitMins = time.Duration(*account.WaitMins)
		}

		feeds = append(feeds, moduleFeed{
			moduleType:   feedTwitterAccount,
			moduleConfig: account,
			waitMins:     waitMins * time.Minute,
		})
	}
}
