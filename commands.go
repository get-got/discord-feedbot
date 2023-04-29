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
			Description: "<FUNCTIONING> Add a new feed.",
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
			Description: "<FUNCTIONING> Add a new feed.",
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
			Description: "<FUNCTIONING> Add a new feed.",
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
				output += fmt.Sprintf("\n• **%s #%d** \"%s\" \t\t_Last ran %s < %d time%s, every %d minute%s >_",
					getFeedTypeName(moduleFeed.moduleType), moduleFeed.moduleSlot+1, moduleFeed.moduleName,
					humanize.Time(moduleFeed.lastRan), moduleFeed.timesRan, ssuff(moduleFeed.timesRan),
					moduleFeed.waitMins, ssuff(moduleFeed.waitMins),
				)
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: output,
				},
			})
		},

		//#region MODULE MANAGEMENT COMMANDS

		// For adding new module feeds
		"instagram-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
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

				var newFeed configModuleInstagramAccount
				newFeed.Destinations = []string{i.ChannelID}
				if opt, ok := optionMap["handle"]; ok {
					newFeed.ID = opt.StringValue()
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				// Doesn't exist
				if existsInstagramConfig(newFeed.Name) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Instagram Account already exists with that name..."},
					})
					return
				}
				if newFeed.Name == "" || newFeed.ID == "" {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Config name or feed identifier was empty... Try again!"},
					})
					return
				}
				if opt, ok := optionMap["wait"]; ok {
					val := int(opt.IntValue())
					newFeed.WaitMins = &val
				}

				feedIndex := len(instagramConfig.Accounts) // cache index for new routine
				instagramConfig.Accounts = append(instagramConfig.Accounts, newFeed)

				err := saveModuleConfig(feedInstagramAccount)
				if err != nil {
					log.Println(color.HiRedString("error saving config")) //TODO:
				} else {
					reply := "Added new Instagram Account! Saved to config..."
					json, err := json.MarshalIndent(newFeed, "", "\t")
					if err == nil {
						reply += fmt.Sprintf("\n```json\n%s```", json)
					}
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: reply},
					})
				}

				waitMins := instagramConfig.WaitMins
				if newFeed.WaitMins != nil {
					waitMins = *newFeed.WaitMins
				}

				feeds = append(feeds, moduleFeed{
					moduleSlot:   feedIndex,
					moduleType:   feedInstagramAccount,
					moduleName:   newFeed.Name,
					moduleRef:    newFeed.ID,
					moduleConfig: newFeed,
					waitMins:     waitMins,
				})
				go startFeed(feedIndex)
			}
		},
		// For modifying existing module feeds
		"instagram-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
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
		// For removing existing module feeds
		"instagram-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
		},

		// For adding new module feeds
		"rss-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
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

				var newFeed configModuleRssFeed
				newFeed.Destinations = []string{i.ChannelID}
				if opt, ok := optionMap["url"]; ok {
					newFeed.URL = opt.StringValue()
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				// Doesn't exist
				if existsRssConfig(newFeed.Name) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "RSS Feed already exists with that name..."},
					})
					return
				}
				if newFeed.Name == "" || newFeed.URL == "" {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Config name or feed identifier was empty... Try again!"},
					})
					return
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

				feedIndex := len(rssConfig.Feeds) // cache index for new routine
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
					waitMins:     waitMins,
				})
				go startFeed(feedIndex)
			}
		},
		// For modifying existing module feeds
		"rss-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
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
		// For removing existing module feeds
		"rss-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
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

					if !existsRssConfig(name) {
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

		// For adding new module feeds
		"twitter-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
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

				var newFeed configModuleTwitterAccount
				newFeed.Destinations = []string{i.ChannelID}
				if opt, ok := optionMap["handle"]; ok {
					//TODO: HANDLE -> ID
					newFeed.ID = opt.StringValue()
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				// Doesn't exist
				if existsTwitterConfig(newFeed.Name) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Twitter Account already exists with that name..."},
					})
					return
				}
				if newFeed.Name == "" || newFeed.ID == "" {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Config name or feed identifier was empty... Try again!"},
					})
					return
				}
				if opt, ok := optionMap["wait"]; ok {
					val := int(opt.IntValue())
					newFeed.WaitMins = &val
				}

				feedIndex := len(twitterConfig.Accounts) // cache index for new routine
				twitterConfig.Accounts = append(twitterConfig.Accounts, newFeed)

				err := saveModuleConfig(feedTwitterAccount)
				if err != nil {
					log.Println(color.HiRedString("error saving config")) //TODO:
				} else {
					reply := "Added new Twitter Account! Saved to config..."
					json, err := json.MarshalIndent(newFeed, "", "\t")
					if err == nil {
						reply += fmt.Sprintf("\n```json\n%s```", json)
					}
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: reply},
					})
				}

				waitMins := twitterConfig.WaitMins
				if newFeed.WaitMins != nil {
					waitMins = *newFeed.WaitMins
				}

				feeds = append(feeds, moduleFeed{
					moduleSlot:   feedIndex,
					moduleType:   feedTwitterAccount,
					moduleName:   newFeed.Name,
					moduleRef:    newFeed.ID,
					moduleConfig: newFeed,
					waitMins:     waitMins,
				})
				go startFeed(feedIndex)
			}
		},
		// For modifying existing module feeds
		"twitter-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
			authorUser := getAuthor(i)
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
		// For removing existing module feeds
		"twitter-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
			authorUser := getAuthor(i)
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

		//#endregion
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
