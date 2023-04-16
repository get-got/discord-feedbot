package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
)

/*
--- PROJECT ---
* Resolve logging standard for now
* Optional build includes?
** Start Settings framework ; base settings & layout
** Start Settings framework ; routine/scheduling/timing structure
** Start Settings framework ; command interactions ; full remote control, headless given errorless
** Start Settings framework ; v module settings , extension settings , etc
*** Start Module framework ;  ^ settings structure
*** Start Module framework ; routine/scheduling/timing structure
*** Start Module framework ; command interactions ; full remote control, headless given errorless
*** Start Module framework ; item cataloging ; db solution

-- LIBS --
* CLI Command Input?
* More discordgo exts?
* Routine exts?
* Color/log alts?

--- MODULES ---
* Twitter Tweets ; Start structuring
* RSS ; Start structuring
* Instagram ; Start structuring
* Spotify Artist Releases ; Start structuring
*M Flickr ; Start structuring
*M System Monitor ; Start structuring
*L Twitter Trends ; Start structuring
*L NASA APOD ; Start structuring
*L Plex Titles ; Start structuring
*L Twitch Chat Track ; Start structuring
*L Twitch Live ; Start structuring
*L Spotify Playlist Changes ; Start structuring

*/

var (

	//TODO:
	// Bot
	/*bot      *discordgo.Session
	botReady bool = false
	user     *discordgo.User
	dgr      *exrouter.Route*/

	// General
	loop         chan os.Signal
	timeLaunched time.Time
)

//TODO:
/*type module struct {
	ref          string        // name
	defaultSleep time.Duration // sleep delay before executing again, TODO: replace with :00 :15 :30 time-based execution?
}

// only used to track execution
type moduleFeed struct { // i.e. thread, account, source, etc. sub of module
	module  *module       // point to parent
	ref     string        // name
	sleep   time.Duration // sleep delay before executing again, TODO: replace with :00 :15 :30 time-based execution?
	lastRan time.Time     // time last ran
}*/

func init() {
	loop = make(chan os.Signal, 1)
	timeLaunched = time.Now()

	//TODO:
	// ensure program has proper permissions

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(color.HiCyanString(wrapHyphensW(fmt.Sprintf("Welcome to %s v%s", projectName, projectVersion))))

	//TODO: Github Update Check

	//TODO: Settings Parse

	//TODO: Settings Create, with defaults if missing

	//TODO: Database Load, create if missing
}

func main() {

	//TODO: Discord Login
	//botLogin()
	//botReady = true

	//TODO: API Logins

	//TODO: Launch Module Managers ?????????????????????????????????????????????
	// start tickers? idk yet

	//TODO:
	//"Startup finished, took %s...", uptime())

	// Infinite loop until interrupted
	signal.Notify(loop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)
	<-loop

	log.Println(color.HiRedString("Exiting... "))
}
