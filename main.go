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
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectLabel, projectVersion))))

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
	databaseError := loadDatabase()
	if databaseError != nil {
		log.Println(color.RedString("loadDatabase() ERROR: %s", databaseError))
	}
}

var (
	instagramAccount_Channel = make(chan feedThread)
	rssFeed_Channel          = make(chan feedThread)
	twitterAccount_Channel   = make(chan feedThread)
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

	// Discord Presence
	if discordConfig.PresenceEnabled {
		go func() {
			for {
				/*
					output += fmt.Sprintf("\n• %s: `%s` \t\t_Last ran %s < %d time%s, every %d minute%s >_",
						getFeedTypeName(feedThread.Group), feedThread.Name,
						humanize.Time(feedThread.LastRan), feedThread.TimesRan, ssuff(feedThread.TimesRan),
						feedThread.WaitMins, ssuff(feedThread.WaitMins),
					)*/
				presence := ""

				// 1st - Link Count
				discord.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name: fmt.Sprintf("%d links stored", refCount()),
						Type: discordgo.ActivityTypeGame,
					}},
					Status: discordConfig.PresenceType,
				})
				time.Sleep(time.Duration(discordConfig.PresenceRefreshRate * int(time.Second)))

				// 2nd - Feed Count
				feedCount := getFeedCount(feed0)
				if feedCount == 0 {
					presence = "no feeds"
				} else if feedCount == 1 {
					presence = "1 feed"
				} else { // 2+
					presence = fmt.Sprintf("%d feeds", feedCount)
				}
				discord.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name: presence,
						Type: discordgo.ActivityTypeListening,
					}},
					Status: discordConfig.PresenceType,
				})
				time.Sleep(time.Duration(discordConfig.PresenceRefreshRate * int(time.Second)))

				// 3rd - Feed Activity
				feedsRunning := getFeedsRunningCount(feed0)
				if feedsRunning == 0 {
					presence = "no feeds running"
				} else if feedsRunning == 1 {
					presence = "1 feed running"
				} else { // 2+
					presence = fmt.Sprintf("%d feeds running", feedsRunning)
				}
				discord.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name: presence,
						Type: discordgo.ActivityTypeWatching,
					}},
					Status: discordConfig.PresenceType,
				})
				time.Sleep(time.Duration(discordConfig.PresenceRefreshRate * int(time.Second)))

				// 4th - Latest Feed
				latestFeed := getFeedsLatest()
				if latestFeed != nil {
					presence = fmt.Sprintf("last feed %s/%s", latestFeed.Name, latestFeed.Ref)
					discord.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{{
							Name: presence,
							Type: discordgo.ActivityTypeCompeting,
						}},
						Status: discordConfig.PresenceType,
					})
					time.Sleep(time.Duration(discordConfig.PresenceRefreshRate * int(time.Second)))
				}
			}
		}()

	} else if discordConfig.PresenceType != string(discordgo.StatusOnline) {
		discord.UpdateStatusComplex(discordgo.UpdateStatusData{
			Status: discordConfig.PresenceType,
		})
	}

	// Spawn Feeds
	catalogFeeds()
	feedsClone := feeds
	for k := range feedsClone {
		go startFeed(&feeds[k])
	}
	go func() {
		for {
			select {
			case instagramAccount_Triggered := <-instagramAccount_Channel:
				{
					if err := handleInstagramAccount(instagramAccount_Triggered.Config.(configModuleInstagramAccount)); err != nil {
						log.Println(color.HiRedString("Error handling Instagram Account: %s", err.Error()))
					}
				}
			case rssFeed_Triggered := <-rssFeed_Channel:
				{
					if err := handleRssFeed(rssFeed_Triggered.Config.(configModuleRssFeed)); err != nil {
						log.Println(color.HiRedString("Error handling RSS Feed: %s", err.Error()))
					}
				}
			case twitterAccount_Triggered := <-twitterAccount_Channel:
				{
					if err := handleTwitterAcc(twitterAccount_Triggered.Config.(configModuleTwitterAcc)); err != nil {
						log.Println(color.HiRedString("Error handling Twitter Account: %s", err.Error()))
					}
				}
			}
			time.Sleep(50 * time.Millisecond) // don't wanna loop infinitely with no delay
		}
	}()

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
