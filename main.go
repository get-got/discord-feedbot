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

	"github.com/bwmarrin/discordgo"
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

func main() {

	if err := openDiscord(); err != nil {
		log.Println(color.HiRedString("DISCORD LOGIN ERROR: %s", err))
	}
	for api, err := range openAPIs() {
		if err != nil {
			log.Println(color.HiRedString("API LOGIN ERROR (%s): %s", api, err))
		}
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	startFeeds()
	// This will need to be heavily modified to allow for live changes to configs
	feedsCopy := feeds
	for key, feed := range feedsCopy {
		go func(key int, feed moduleFeed) {
			for {
				feeds[key].timesRan++
				feeds[key].lastRan = time.Now()
				//TODO: Error handling
				switch feed.moduleType {
				case feedInstagramAccount:
					{
						go func() {
							if err := handleInstagramAccount(feed.moduleConfig.(configModuleInstagramAccount)); err != nil {
								log.Println(color.HiRedString("Error handling Instagram Account: %s", err.Error()))
							}
						}()
					}
				case feedTwitterAccount:
					{
						go func() {
							if err := handleTwitterAccount(feed.moduleConfig.(configModuleTwitterAccount)); err != nil {
								log.Println(color.HiRedString("Error handling Twitter Account: %s", err.Error()))
							}
						}()
					}
				case feedRSS_Feed:
					{
						go func() {
							if err := handleRSS_Feed(feed.moduleConfig.(configModuleRSS_Feed)); err != nil {
								log.Println(color.HiRedString("Error handling RSS Feed: %s", err.Error()))
							}
						}()
					}
				}
				time.Sleep(feed.waitMins)
			}
		}(key, feed)
	}

	if generalConfig.Debug {
		log.Println(color.HiYellowString("Startup finished, took %s...", uptime()))
	}

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := discord.ApplicationCommandDelete(discord.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
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

const (
	feed000 = iota

	feedInstagramAccount
	feedFlickrGroup
	feedFlickrUser
	feedRSS_Feed
	feedSpotifyArtist
	feedSpotifyPlaylist
	feedSpotifyPodcast
	feedTwitterAccount
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
