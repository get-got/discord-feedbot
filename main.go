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

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix)
	log.SetPrefix("| ")
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))

	//TODO: Github Update Check

	// Load Configs
	settingsErrors := loadConfig()
	if len(settingsErrors) > 0 {
		log.Println(color.HiRedString("loadConfig(): Detected errors in settings..."))
		for section, err := range settingsErrors {
			if err != nil {
				log.Println(color.RedString("loadConfig()[\"%s\"] ERROR: %s", section, err))
			}
		}
	}

	//TODO: Database Load, create if missing
	loadDatabase()
}

var (
	instagramAccount_Channel = make(chan moduleFeed)
	rssFeed_Channel          = make(chan moduleFeed)
	twitterAccount_Channel   = make(chan moduleFeed)
)

func main() {

	if err := openDiscord(); err != nil {
		log.Println(color.HiRedString("DISCORD LOGIN ERROR: %s", err))
	}
	go addSlashCommands()
	for api, err := range openAPIs() {
		if err != nil {
			log.Println(color.HiRedString("API LOGIN ERROR (%s): %s", api, err))
		}
	}

	if generalConfig.Debug {
		log.Println(color.HiYellowString("Startup finished, took %s...", uptime()))
	}

	// Spawn Feeds
	indexFeeds()
	feedsClone := feeds
	for k := range feedsClone {
		go startFeed(&feeds[k])
	}
	for {
		select {
		case instagramAccount_Triggered := <-instagramAccount_Channel:
			{
				feed := instagramAccount_Triggered
				log.Println(color.HiMagentaString("instagram %s fired", feed.Name))
				if err := handleInstagramAccount(feed.Config.(configModuleInstagramAccount)); err != nil {
					log.Println(color.HiRedString("Error handling Instagram Account: %s", err.Error()))
				}
			}
		case rssFeed_Triggered := <-rssFeed_Channel:
			{
				feed := rssFeed_Triggered
				log.Println(color.HiMagentaString("rss %s fired", feed.Name))
				if err := handleRssFeed(feed.Config.(configModuleRssFeed)); err != nil {
					log.Println(color.HiRedString("Error handling RSS Feed: %s", err.Error()))
				}
			}
		case twitterAccount_Triggered := <-twitterAccount_Channel:
			{
				feed := twitterAccount_Triggered
				log.Println(color.HiMagentaString("twitter %s fired", feed.Name))
				if err := handleTwitterAccount(feed.Config.(configModuleTwitterAccount)); err != nil {
					log.Println(color.HiRedString("Error handling Twitter Account: %s", err.Error()))
				}
			}
		}
		time.Sleep(50 * time.Millisecond) // don't wanna loop infinitely with no delay
	}

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	if discordConfig.DeleteCommands {
		deleteSlashCommands()
	}

	log.Println(color.GreenString("Logging out of discord..."))
	discord.Close()

	log.Println(color.HiRedString("Exiting... "))
}

func openAPIs() map[string]error {
	errors := make(map[string]error)
	var tmperr error

	if tmperr = openInstagram(); tmperr != nil {
		errors["login-instagram"] = tmperr
	}

	if tmperr = openTwitter(); tmperr != nil {
		errors["login-twitter"] = tmperr
	}

	return errors
}
