package main

import (
	"log"

	"github.com/fatih/color"
)

/* TODO:

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
* Flickr ; Start structuring
* Instagram ; Start structuring
* NASA APOD ; Start structuring
* Overwatch Patchnotes ; Start structuring
* Plex Titles ; Start structuring
* RSS ; Start structuring
* Spotify Artist Releases ; Start structuring
* Spotify Playlist Changes ; Start structuring
* System Monitor ; Start structuring
* Twitch Chat Track ; Start structuring
* Twitch Live ; Start structuring
* Twitter Tweets ; Start structuring
* Twitter Trends ; Start structuring

*/

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(color.Output)
	log.Println(">>> init() ...")
	//
	log.Println("<<< init() ...")
}

func main() {
	log.Println(">>> main() ...")
	//
	log.Println("<<< main() ...")
}
