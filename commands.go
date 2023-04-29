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
			Name:        "ping",
			Description: "Simple ping pong latency test.",
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
			Name:        "feeds",
			Description: "<WIP> General overview of the bot.",
		},

		{
			Name:        "instagram-add",
			Description: "<WIP> Add a new feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Instagram Account ID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "wait",
					Description: "Feed Delay (x Minutes)",
					Required:    false,
				},
			},
		},
		{
			Name:        "instagram-modify",
			Description: "<WIP> Modify an existing feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
			},
		},
		{
			Name:        "instagram-delete",
			Description: "<WIP> Delete an existing feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
			},
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
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "wait",
					Description: "Feed Delay (x Minutes)",
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
					Name:        "username",
					Description: "Webhook Username",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "avatar",
					Description: "Webhook Avatar Image URL",
					Required:    false,
				},
				/*{
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
				},*/
			},
		},
		{
			Name:        "rss-modify",
			Description: "<WIP> Modify an existing feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
			},
		},
		{
			Name:        "rss-delete",
			Description: "<WIP> Delete an existing feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
			},
		},

		{
			Name:        "twitter-add",
			Description: "<WIP> Add a new feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "Twitter Account ID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "wait",
					Description: "Feed Delay (x Minutes)",
					Required:    false,
				},
			},
		},
		{
			Name:        "twitter-modify",
			Description: "<WIP> Modify an existing feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
			},
		},
		{
			Name:        "twitter-delete",
			Description: "<WIP> Delete an existing feed.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Unique Feed Name",
					Required:    true,
				},
			},
		},
	}

	commandNotAdmin = "Your Discord ID must be listed as an admin in the `discord.json` settings for this bot to use this command."

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			beforePong := time.Now()
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong!",
				},
			})
			if err == nil {
				afterPong := time.Now()
				latency := discord.HeartbeatLatency().Milliseconds()
				roundtrip := afterPong.Sub(beforePong).Milliseconds()
				content := fmt.Sprintf("**Latency:** ``%dms`` — **Roundtrip:** ``%dms``",
					latency,
					roundtrip,
				)
				//TODO: embed?
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &content,
				})
			}
		},
		"info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			output := ""
			for component, version := range getComponentVersions() {
				output += fmt.Sprintf("\n%s %s", component, version)
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: output,
				},
			})
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			/*json, _ := json.MarshalIndent(moduleFeed.moduleConfig, "", "\t")
			output += fmt.Sprintf("\n• **%s #%d** \t\t_Last ran %s < %d time%s, every %d minute%s >_\n```json\n%s```",
				getFeedTypeName(moduleFeed.moduleType), moduleFeed.moduleSlot+1,
				humanize.Time(moduleFeed.lastRan), moduleFeed.timesRan, ssuff(moduleFeed.timesRan),
				waitMins, ssuff(waitMins), json,
			)*/
		},
		"feeds": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			output := "**ACTIVE FEEDS:**\n"
			for _, moduleFeed := range feeds {
				waitMins := int(moduleFeed.waitMins / time.Minute)
				output += fmt.Sprintf("\n• **%s #%d** \t\t_Last ran %s < %d time%s, every %d minute%s >_",
					getFeedTypeName(moduleFeed.moduleType), moduleFeed.moduleSlot+1, //disableLinks(moduleFeed.moduleRef),
					humanize.Time(moduleFeed.lastRan), moduleFeed.timesRan, ssuff(moduleFeed.timesRan),
					waitMins, ssuff(waitMins),
				)
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: output,
				},
			})
		},

		// MODULE MANAGEMENT COMMANDS

		"instagram-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For adding new module feeds
			//TODO: everything
		},
		"instagram-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For modifying existing module feeds
			var authorUser *discordgo.User
			if i.Member == nil && i.Message.Author == nil {
				return
			}
			if i.Member != nil {
				authorUser = i.Member.User
			} else if i.Message.Author != nil {
				authorUser = i.Message.Author
			}
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
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
				//TODO: everything
			}
		},
		"instagram-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For removing existing module feeds
			//TODO: everything
		},

		"rss-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For adding new module feeds

			var authorUser *discordgo.User
			if i.Member == nil && i.Message.Author == nil {
				return
			}
			if i.Member != nil {
				authorUser = i.Member.User
			} else if i.Message.Author != nil {
				authorUser = i.Message.Author
			}
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
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
				if opt, ok := optionMap["avatar"]; ok {
					val := opt.StringValue()
					newFeed.Avatar = &val
				}
				if opt, ok := optionMap["username"]; ok {
					val := opt.StringValue()
					newFeed.Username = &val
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

				if existsRSSFeed(newFeed.Name) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "RSS Feed already exists with that name..."},
					})
					return
				}

				feedIndex := len(rssConfig.Feeds)
				rssConfig.Feeds = append(rssConfig.Feeds, newFeed)

				err := saveModuleConfig(feedRSS)
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

				waitMins := rssConfig.WaitMins
				if newFeed.WaitMins != nil {
					waitMins = *newFeed.WaitMins
				}

				feeds = append(feeds, moduleFeed{
					moduleSlot:   feedIndex,
					moduleType:   feedRSS,
					moduleName:   newFeed.Name,
					moduleRef:    "\"" + newFeed.URL + "\"",
					moduleConfig: newFeed,
					waitMins:     time.Duration(waitMins),
				})
				go startFeed(feedIndex)
			}
		},
		"rss-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For modifying existing module feeds
			var authorUser *discordgo.User
			if i.Member == nil && i.Message.Author == nil {
				return
			}
			if i.Member != nil {
				authorUser = i.Member.User
			} else if i.Message.Author != nil {
				authorUser = i.Message.Author
			}
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
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
				//TODO: everything
			}
		},
		"rss-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For removing existing module feeds
			var authorUser *discordgo.User
			if i.Member == nil && i.Message.Author == nil {
				return
			}
			if i.Member != nil {
				authorUser = i.Member.User
			} else if i.Message.Author != nil {
				authorUser = i.Message.Author
			}
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
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

				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()

					if !existsRSSFeed(name) {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{Content: "No RSS Feed exists with that name..."},
						})
						return
					} else {
						//TODO: everything
						// probably need to kill the go routine before deleting from config?
						/*s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{Content: "RSS Feed already exists with that name..."},
						})*/
					}
				}
			}
		},

		"twitter-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For adding new module feeds
			//TODO: everything
		},
		"twitter-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For modifying existing module feeds
			var authorUser *discordgo.User
			if i.Member == nil && i.Message.Author == nil {
				return
			}
			if i.Member != nil {
				authorUser = i.Member.User
			} else if i.Message.Author != nil {
				authorUser = i.Message.Author
			}
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
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
				//TODO: everything
			}
		},
		"twitter-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// For removing existing module feeds
			//TODO: everything
		},
	}
)

var slashCommands []*discordgo.ApplicationCommand

func addSlashCommands() {
	log.Println(color.CyanString("Initializing slash commands...\tCommands won't work until this finishes..."))
	slashCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create command '%v': %v", v.Name, err)
		}
		slashCommands[i] = cmd
	}
	log.Println(color.HiCyanString("Slash commands created!\tYou can now use Discord commands..."))
}

func deleteSlashCommands() {
	log.Println(color.CyanString("Deleting slash commands...\tThis may take a bit..."))
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
	case feedInstagramAccount:
		return saveConfig(pathConfigModuleInstagram, instagramConfig)
	case feedRSS:
		return saveConfig(pathConfigModuleRSS, rssConfig)
	case feedTwitterAccount:
		return saveConfig(pathConfigModuleTwitter, twitterConfig)
	}
	return nil
}
