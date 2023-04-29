package main

import (
	"log"
	"time"

	"github.com/fatih/color"
)

type moduleFeed struct { // i.e. thread, account, source, etc. sub of module
	moduleSlot   int
	moduleType   int
	moduleName   string
	moduleRef    string
	moduleConfig interface{} // point to parent
	waitMins     time.Duration
	lastRan      time.Time
	timesRan     int
}

var feeds []moduleFeed

const (
	feed000 = iota

	feedInstagramAccount
	feedFlickrGroup
	feedFlickrUser
	feedRSS
	feedSpotifyArtist
	feedSpotifyPlaylist
	feedSpotifyPodcast
	feedTwitterAccount
)

func getFeedTypeName(moduleType int) string {
	switch moduleType {
	case feedInstagramAccount:
		return "Instagram Account"
	case feedFlickrGroup:
		return "Flickr Group"
	case feedFlickrUser:
		return "Flickr User"
	case feedRSS:
		return "RSS Feed"
	case feedSpotifyArtist:
		return "Spotify Artist"
	case feedSpotifyPlaylist:
		return "Spotify Playlist"
	case feedSpotifyPodcast:
		return "Spotify Podcast"
	case feedTwitterAccount:
		return "Twitter Account"
	}
	return ""
}

func indexFeeds() {
	// RSS Feeds
	for k, feed := range rssConfig.Feeds {
		waitMins := time.Duration(rssConfig.WaitMins)
		if feed.WaitMins != nil {
			waitMins = time.Duration(*feed.WaitMins)
		}

		feeds = append(feeds, moduleFeed{
			moduleSlot:   k,
			moduleType:   feedRSS,
			moduleName:   feed.Name,
			moduleRef:    "\"" + feed.URL + "\"",
			moduleConfig: feed,
			waitMins:     waitMins,
		})
	}
	// Instagram, Accounts
	for k, account := range instagramConfig.Accounts {
		waitMins := time.Duration(instagramConfig.WaitMins)
		if account.WaitMins != nil {
			waitMins = time.Duration(*account.WaitMins)
		}

		feeds = append(feeds, moduleFeed{
			moduleSlot:   k,
			moduleType:   feedInstagramAccount,
			moduleName:   account.Name,
			moduleRef:    account.ID,
			moduleConfig: account,
			waitMins:     waitMins,
		})
	}
	// Twitter, Accounts
	for k, account := range twitterConfig.Accounts {
		waitMins := twitterConfig.WaitMins
		if account.WaitMins != nil {
			waitMins = *account.WaitMins
		}

		feeds = append(feeds, moduleFeed{
			moduleSlot:   k,
			moduleType:   feedTwitterAccount,
			moduleName:   account.Name,
			moduleRef:    account.ID,
			moduleConfig: account,
			waitMins:     time.Duration(waitMins),
		})
	}
}

func startFeed(key int) {
	for {
		feed := &feeds[key]
		feed.timesRan++
		feed.lastRan = time.Now()
		switch feed.moduleType {
		case feedInstagramAccount:
			{
				if err := handleInstagramAccount(feed.moduleConfig.(configModuleInstagramAccount)); err != nil {
					log.Println(color.HiRedString("Error handling Instagram Account: %s", err.Error()))
				}
			}
		case feedRSS:
			{
				if err := handleRSS_Feed(feed.moduleConfig.(configModuleRSS_Feed)); err != nil {
					log.Println(color.HiRedString("Error handling RSS Feed: %s", err.Error()))
				}
			}
		case feedTwitterAccount:
			{
				if err := handleTwitterAccount(feed.moduleConfig.(configModuleTwitterAccount)); err != nil {
					log.Println(color.HiRedString("Error handling Twitter Account: %s", err.Error()))
				}
			}
		}
		time.Sleep(feed.waitMins * time.Minute)
	}
}
