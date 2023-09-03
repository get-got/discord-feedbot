package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type feedDestination struct {
	Channel string   `json:"channel"`
	Tags    []string `json:"tags,omitempty"`
}

type feedThread struct {
	Group    int
	Name     string
	Ref      string
	Config   interface{} // point to parent
	WaitMins int
	LastRan  time.Time
	TimesRan int
	Running  bool
}

var feeds []feedThread

const (
	feed0 = iota
	feed1

	feedInstagramAccount

	//feedFlickrGroup
	//feedFlickrUser

	feedRSS

	//feedSpotifyArtist
	//feedSpotifyPlaylist
	//feedSpotifyPodcast

	feedTwitterAccount
)

func getFeedTypeName(moduleType int) string {
	switch moduleType {
	case feed0:
		return ""
	case feed1:
		return "PLACEHOLDER"
	case feedInstagramAccount:
		return "Instagram Account"
	//case feedFlickrGroup:
	//	return "Flickr Group"
	//case feedFlickrUser:
	//	return "Flickr User"
	case feedRSS:
		return "RSS Feed"
	//case feedSpotifyArtist:
	//	return "Spotify Artist"
	//case feedSpotifyPlaylist:
	//	return "Spotify Playlist"
	//case feedSpotifyPodcast:
	//	return "Spotify Podcast"
	case feedTwitterAccount:
		return "Twitter Account"
	}
	return ""
}

func getFeedCount(filterGroup int) int {
	if filterGroup != feed0 {
		counter := 0
		for _, feed := range feeds {
			if feed.Group == filterGroup {
				counter++
			}
		}
		return counter
	}
	return len(feeds)
}

func getFeedsRunningCount(filterGroup int) int {
	counter := 0
	for _, feed := range feeds {
		if feed.Running && feed.Group == filterGroup {
			counter++
		}
	}
	return counter
}

func getFeedsSleepingCount(filterGroup int) int {
	counter := 0
	for _, feed := range feeds {
		if !feed.Running && feed.Group == filterGroup {
			counter++
		}
	}
	return counter
}

func getFeedsLatest() *feedThread {
	var latestFeed *feedThread
	for _, feed := range feeds {
		if feed.LastRan.After(latestFeed.LastRan) {
			latestFeed = &feed
		}
	}
	return latestFeed
}

func catalogFeeds() {
	//feeds = make([]moduleFeed, 0)
	// RSS Feeds
	for _, feed := range rssConfig.Feeds {
		waitMins := rssConfig.WaitMins
		if feed.WaitMins != nil {
			waitMins = *feed.WaitMins
		}

		feeds = append(feeds, feedThread{
			Group:    feedRSS,
			Name:     feed.Name,
			Ref:      "\"" + feed.URL + "\"",
			Config:   feed,
			WaitMins: waitMins,
		})
	}
	// Instagram, Accounts
	for _, account := range instagramConfig.Accounts {
		waitMins := instagramConfig.WaitMins
		if account.WaitMins != nil {
			waitMins = *account.WaitMins
		}

		feeds = append(feeds, feedThread{
			Group:    feedInstagramAccount,
			Name:     account.Name,
			Ref:      account.ID,
			Config:   account,
			WaitMins: waitMins,
		})
	}
	// Twitter, Accounts
	for _, account := range twitterConfig.Accounts {
		waitMins := twitterConfig.WaitMins
		if account.WaitMins != nil {
			waitMins = *account.WaitMins
		}

		feeds = append(feeds, feedThread{
			Group:    feedTwitterAccount,
			Name:     account.Name,
			Ref:      account.Handle,
			Config:   account,
			WaitMins: waitMins,
		})
	}
}

func startFeed(feed *feedThread) {
	for {
		if feed == nil { // deleted
			break
		}
		if feed.Name == "" || feed.Ref == "" { // deleted
			break
		}
		feed.TimesRan++
		feed.LastRan = time.Now()
		feed.Running = true
		switch feed.Group {
		case feedInstagramAccount:
			{
				instagramAccount_Channel <- *feed
			}
		case feedRSS:
			{
				rssFeed_Channel <- *feed
			}
		case feedTwitterAccount:
			{
				twitterAccount_Channel <- *feed
			}
		}
		feed.Running = false
		time.Sleep(time.Duration(feed.WaitMins * int(time.Minute)))
	}
}

func getModuleFeed(name string, group int) *feedThread {
	for k, feed := range feeds {
		if feed.Name == name && feed.Group == group {
			return &feeds[k]
		}
	}
	return nil
}

func updateFeedConfig(name string, group int, config interface{}) bool {
	cloneFeeds := feeds
	for i, feed := range cloneFeeds {
		if feed.Name == name && feed.Group == group {
			feeds[i].Config = config
			return true
		}
	}
	return false
}

func deleteFeed(name string, group int) bool {
	cloneFeeds := feeds
	for i, feed := range cloneFeeds {
		if feed.Name == name && feed.Group == group {
			feeds = append(feeds[:i], feeds[i+1:]...)
			return true
		}
	}
	return false
}

func saveModuleConfig(feedType int) error {
	switch feedType {
	case feedInstagramAccount:
		return saveConfig(pathConfigModuleInstagram, instagramConfig)
	case feedRSS:
		return saveConfig(pathConfigModuleRSS, rssConfig)
	case feedTwitterAccount:
		return saveConfig(pathConfigModuleTwitter, twitterConfig)
	}
	return nil
}

func replyConfig(jsonFeed interface{}, reply string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	var jsonn []byte
	var err error
	if jsonn, err = json.MarshalIndent(jsonFeed, "", "\t"); err == nil {
		reply += fmt.Sprintf("\n```json\n%s```", jsonn)
	} else {
		return err
	}
	InteractionRespond(reply, s, i)
	return nil
}

func saveModuleConfigReply(moduleGroup int, jsonFeed interface{}, reply string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := saveModuleConfig(moduleGroup); err != nil { // save config
		return fmt.Errorf("error saving config: %s", err.Error())
	} else {
		return replyConfig(jsonFeed, reply, s, i)
	}
}
