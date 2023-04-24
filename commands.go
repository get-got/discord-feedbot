package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

/*

/config main
>>> spits out current config
>>> buttons to edit or reset

*/

var (
	// https://github.com/bwmarrin/discordgo/blob/master/examples/slash_commands/main.go
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "<WIP>",
		},
		{
			Name:        "info",
			Description: "<WIP> General info about this project.",
		},
		{
			Name:        "status",
			Description: "<WIP> General overview of the bot.",
		},
		{
			Name:        "ping",
			Description: "Simple ping pong latency test.",
		},

		{
			Name:        "rss-add",
			Description: "<WIP> Add a new feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "RSS Feed URL",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "wait",
					Description: "RSS Feed Delay (x Minutes)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "RSS Feed Name",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "twitter",
					Description: "Twitter for Username & Avatar",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "txt",
					Description: "xxz",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "aa",
							Value: "aa",
						},
						{
							Name:  "bb",
							Value: "bb",
						},
						{
							Name:  "cc",
							Value: "cc",
						},
					},
				},
			},
		},
		{
			Name:        "rss-modify",
			Description: "<WIP> Modify an existing feed.",
		},
		{
			Name:        "rss-delete",
			Description: "<WIP> Delete an existing feed.",
		},

		{
			Name:        "twitter-add",
			Description: "<WIP> Add a new feed.",
		},
		{
			Name:        "twitter-modify",
			Description: "<WIP> Modify an existing feed.",
		},
		{
			Name:        "twitter-delete",
			Description: "<WIP> Delete an existing feed.",
		},
	}

	commandNotAdmin = "Your Discord ID must be listed as an admin in the `discord.json` settings for this bot to use this command."

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		},
		"info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			output := ""
			for component, version := range getComponentVersions() {
				output += fmt.Sprintf("\n%s %s", component, version)
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: output,
				},
			})
			if err != nil {
				//TODO: error handling
			}
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			output := "**ACTIVE FEEDS:**"
			for _, moduleFeed := range feeds {
				waitMins := int(moduleFeed.waitMins / time.Minute)
				output += fmt.Sprintf("\n- **%s #%d** %s \t\t_Last ran %s (%d time%s, every %d minute%s)_",
					getFeedTypeName(moduleFeed.moduleType), moduleFeed.moduleSlot+1,
					disableLinks(moduleFeed.moduleRef),
					humanize.Time(moduleFeed.lastRan), moduleFeed.timesRan, ssuff(moduleFeed.timesRan),
					waitMins, ssuff(waitMins),
				)
			}
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: output,
				},
			})
			if err != nil {
				//TODO: error handling
			}
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			beforePong := time.Now()
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
			if err != nil {
				//TODO: error handling
			} else {
				afterPong := time.Now()
				latency := discord.HeartbeatLatency().Milliseconds()
				roundtrip := afterPong.Sub(beforePong).Milliseconds()
				content := fmt.Sprintf("**Latency:** ``%dms`` — **Roundtrip:** ``%dms``",
					latency,
					roundtrip,
				)
				//TODO: embed?
				_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
				if err != nil {
					//TODO: error handling
				}
			}
		},

		// MODULE MANAGEMENT COMMANDS

		"rss-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For adding new module feeds

			if i.Member == nil { //TODO:
				log.Println(color.HiRedString("ERROR: nil member on command"))
				return
			}

			if !isBotAdmin(i.Member.User.ID) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: commandNotAdmin},
				})
			} else {
				options := i.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}
				//
				var newFeed configModuleRSS_Feed
				newFeed.Destinations = []string{i.ChannelID}
				if opt, ok := optionMap["url"]; ok {
					newFeed.URL = opt.StringValue()
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				if opt, ok := optionMap["wait"]; ok {
					val := int(opt.IntValue())
					newFeed.WaitMins = &val
				}
				if opt, ok := optionMap["twitter"]; ok {
					//TODO: better error checks
					if twitterClient != nil {
						handle := opt.StringValue()
						userResults, err := twitterClient.GetUsersLookup(handle, url.Values{})
						if err != nil {
							if len(userResults) > 0 {
								userResult := userResults[0]
								newFeed.UseTwitter = &userResult.IdStr
							}
						}
					}
				}
				rssConfig.Feeds = append(rssConfig.Feeds, newFeed)
				err := saveModuleConfig(feedRSS_Feed)
				if err != nil {
					log.Println(color.HiRedString("error saving config")) //TODO:
				} else {
					reply := "Added new RSS Feed! Saved to config..."
					json, err := json.MarshalIndent(newFeed, "", "\t")
					if err == nil {
						reply += fmt.Sprintf("\n```json\n%s```", json)
					}
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: reply},
					})
				}

				/*waitMins := time.Duration(rssConfig.WaitMins)
				if newFeed.WaitMins != nil {
					waitMins = time.Duration(*newFeed.WaitMins)
				}*/
				//TODO: Catalog and spin up feed, *** currently requires restart ***
			}
		},
		"rss-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For modifying existing module feeds
		},
		"rss-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For removing existing module feeds
		},

		"twitter-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For adding new module feeds
		},
		"twitter-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For modifying existing module feeds
		},
		"twitter-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For removing existing module feeds
		},
	}
)

var slashCommands []*discordgo.ApplicationCommand

func addSlashCommands() {
	log.Println(color.CyanString("Creating slash commands..."))
	slashCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		slashCommands[i] = cmd
	}
	log.Println(color.HiCyanString("Slash commands created!"))
}

func clearSlashCommands() {
	log.Println(color.CyanString("Deleting slash commands..."))
	for _, v := range slashCommands {
		err := discord.ApplicationCommandDelete(discord.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
	log.Println(color.HiCyanString("Slash commands deleted!"))
}

func saveModuleConfig(feedType int) error {
	switch feedType {
	case feedRSS_Feed:
		return saveConfig(pathConfigModuleRSS, rssConfig)
	case feedTwitterAccount:
		return saveConfig(pathConfigModuleTwitter, twitterConfig)
	}
	return nil
}

func addModuleFeed() {

}

func modifyModuleFeed() {

}

func deleteModuleFeed() {

}
