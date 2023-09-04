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
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
)

var (
	// General
	loop         chan os.Signal
	timeLaunched time.Time
)

func init() {
	loop = make(chan os.Signal, 1)
	timeLaunched = time.Now()

	//#region Initialize Logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix)
	log.SetOutput(color.Output) // TODO: log to file option
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectLabel, projectVersion))))
	l := logInstructions{
		Location: "INITITALIZATION",
		Task:     "entry tasks",
		Inline:   false,
		Color:    color.CyanString,
	}
	//#endregion

	//TODO: Github Update Check

	// Load Configs
	l.Task = "loadConfig"
	settingsErrors := loadConfig()
	if len(settingsErrors) > 0 {
		log.Println(l.SetFlag(&lError).Log("Detected errors in settings..."))
		l.Clear()
		for section, err := range settingsErrors {
			if err != nil {
				log.Println(l.SetFlag(&lError).Log("\"%s\" ERROR: %s", section, err))
				l.Clear()
			}
		}
	}

	// Load Database
	l.Task = "loadDatabase"
	if databaseError := loadDatabase(); databaseError != nil {
		log.Println(l.SetFlag(&lError).Log("ERROR: %s", databaseError))
		l.Clear()
	}
}

var (
	instagramAccount_Channel = make(chan feedThread)
	rssFeed_Channel          = make(chan feedThread)
	twitterAccount_Channel   = make(chan feedThread)
)

func getFeedCountLabel(filterGroup int) string {
	feedCount := getFeedCount(filterGroup)
	if feedCount == 0 {
		return "no feeds"
	} else if feedCount == 1 {
		return "1 feed"
	} else { // 2+
		return fmt.Sprintf("%d feeds", feedCount)
	}
}

func dataKeyReplacement(input string) string {
	//TODO: Case-insensitive key replacement. -- If no streamlined way to do it, convert to lower to find substring location but replace normally
	if strings.Contains(input, "{{") && strings.Contains(input, "}}") {
		timeNow := time.Now()
		keys := [][]string{
			{"{{goVersion}}", runtime.Version()},
			{"{{dgVersion}}", discordgo.VERSION},
			{"{{dfbVersion}}", projectVersion},
			{"{{apiVersion}}", discordgo.APIVersion},
			{"{{numServers}}", fmt.Sprint(len(discord.State.Guilds))},
			{"{{numAdmins}}", fmt.Sprint(len(discordConfig.Admins))},
			{"{{timeNowShort}}", timeNow.Format("3:04pm")},
			{"{{timeNowShortTZ}}", timeNow.Format("3:04pm MST")},
			{"{{timeNowMid}}", timeNow.Format("3:04pm MST 1/2/2006")},
			{"{{timeNowLong}}", timeNow.Format("3:04:05pm MST - January 2, 2006")},
			{"{{timeNowShort24}}", timeNow.Format("15:04")},
			{"{{timeNowShortTZ24}}", timeNow.Format("15:04 MST")},
			{"{{timeNowMid24}}", timeNow.Format("15:04 MST 2/1/2006")},
			{"{{timeNowLong24}}", timeNow.Format("15:04:05 MST - 2 January, 2006")},
			{"{{uptime}}", durafmt.ParseShort(time.Since(timeLaunched)).String()},

			{"{{linkCount}}", fmt.Sprint(refCount())},
			{"{{feedCount}}", string(getFeedCountLabel(feed0))},
		}
		for _, key := range keys {
			if strings.Contains(input, key[0]) {
				input = strings.ReplaceAll(input, key[0], key[1])
			}
		}
	}
	return input
}

func main() {
	l := logInstructions{
		Location: "MAIN",
		Task:     "startup",
		Inline:   false,
		Color:    color.GreenString,
	}

	if err := openDiscord(); err != nil {
		log.Println(l.SetFlag(&lError).Log("DISCORD LOGIN ERROR: %s", err))
		l.ClearFlag()
	}
	if discordConfig.DeleteCommands {
		deleteSlashCommands()
	}
	go addSlashCommands()
	for api, err := range openAPIs() {
		if err != nil {
			log.Println(l.SetFlag(&lError).Log("API LOGIN ERROR: (%s): %s", api, err))
			l.ClearFlag()
		}
	}

	if generalConfig.Verbose {
		log.Println(l.SetFlag(&lVerbose).Log("Startup finished, took %s...", uptime()))
		l.ClearFlag()
	}

	// Start Presence Loop
	if discordConfig.Presence != nil && len(discordConfig.Presence) > 0 {
		go func() {
			for {
				runDiscordPresences() // no need to sleep because the function does after each rotation.
			}
		}()
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
						log.Println(l.SetTask("handleInstagramAccount").SetFlag(&lError).Log(
							"Error handling Instagram Account: %s", err.Error()))
						l.Clear()
					}
				}
			case rssFeed_Triggered := <-rssFeed_Channel:
				{
					if err := handleRssFeed(rssFeed_Triggered.Config.(configModuleRssFeed)); err != nil {
						log.Println(l.SetTask("handleRssFeed").SetFlag(&lError).Log(
							"Error handling RSS Feed: %s", err.Error()))
						l.Clear()
					}
				}
			case twitterAccount_Triggered := <-twitterAccount_Channel:
				{
					if err := handleTwitterAcc(twitterAccount_Triggered.Config.(configModuleTwitterAcc)); err != nil {
						log.Println(l.SetTask("handleTwitterAcc").SetFlag(&lError).Log(
							"Error handling Twitter Account: %s", err.Error()))
						l.Clear()
					}
				}
			}
			time.Sleep(50 * time.Millisecond) // don't wanna loop infinitely with no delay
		}
	}()

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	l.Task = "exit"

	if discordConfig.DeleteCommands {
		deleteSlashCommands()
	}

	log.Println(l.Log("Logging out of discord..."))
	discord.Close()

	log.Println(l.LogC(color.HiRedString, "Exiting..."))
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
