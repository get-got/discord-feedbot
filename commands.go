package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
)

var (
	nameCommand = []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "Unique Feed Name",
			Required:    true,
		},
	}

	genericModuleCommands = []*discordgo.ApplicationCommandOption{
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
			Type:        discordgo.ApplicationCommandOptionMentionable,
			Name:        "tag",
			Description: "Add Discord user to tag",
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
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "blacklist",
			Description: "Add Blacklist Line (sep by \"|\")",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "whitelist",
			Description: "Add Whitelist Line (sep by \"|\")",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "list-type",
			Description: "Ordering for using Blacklist AND Whitelist together",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Black then White (Default)",
					Value: "bw",
				},
				{
					Name:  "White then Black",
					Value: "wb",
				},
			},
		},
	}

	// https://github.com/bwmarrin/discordgo/blob/master/examples/slash_commands/main.go
	commands = []*discordgo.ApplicationCommand{

		//#region General
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
		//#endregion

		//#region RSS Feeds
		{
			Name:        "rss-new",
			Description: "Add a new feed",
			Options: append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "RSS Feed URL",
				Required:    true,
			}}, genericModuleCommands...),
		},
		{
			Name:        "rss-add",
			Description: "Add this channel to an existing feed",
			Options:     nameCommand,
		},
		{
			Name:        "rss-modify",
			Description: "Modify an existing feed",
			Options: append(genericModuleCommands, []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "change-url",
					Description: "Modify RSS Feed URL",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "blacklist-url",
					Description: "Add Blacklist Line (sep by \"|\") for URL contents",
					Required:    false,
				},
			}...),
		},
		{
			Name:        "rss-delete",
			Description: "Delete an existing feed",
			Options:     nameCommand,
		},
		{
			Name:        "rss-show",
			Description: "Display info for an existing feed",
			Options:     nameCommand,
		},
		//#endregion

		//#region Twitter Accounts
		{
			Name:        "twitter-new",
			Description: "Add a new feed",
			Options: append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "handle",
				Description: "Twitter Handle (@)",
				Required:    true,
			}}, genericModuleCommands...),
		},
		{
			Name:        "twitter-add",
			Description: "Add this channel to an existing feed",
			Options:     nameCommand,
		},
		{
			Name:        "twitter-modify",
			Description: "Modify an existing feed",
			Options: append(genericModuleCommands, []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "change-handle",
					Description: "Change Twitter Handle (@)",
					Required:    false,
				},
			}...),
		},
		{
			Name:        "twitter-delete",
			Description: "Delete an existing feed",
			Options:     nameCommand,
		},
		{
			Name:        "twitter-show",
			Description: "Display info for an existing feed",
			Options:     nameCommand,
		},
		//#endregion

	}

	commandNotAdmin = "Your Discord ID must be listed as an admin in the `discord.json` settings for this bot to use this command."

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "wip"},
			})
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
			output := "**" + projectName + " " + projectVersion + "**"
			output += "\n• discordgo v" + discordgo.VERSION + " with Discord API v" + discordgo.APIVersion
			output += "\n• Twitter API v1.1"
			output += "\n_Launched " + humanize.Time(timeLaunched) + "_"
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: output,
				},
			})
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "wip"},
			})
			//TODO: everything
		},
		"feeds": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			output := ""
			for _, moduleFeed := range feeds {
				output += fmt.Sprintf("\n• %s: `%s` \t\t_Last ran %s < %d time%s, every %d minute%s >_",
					getFeedTypeName(moduleFeed.Group), moduleFeed.Name,
					humanize.Time(moduleFeed.LastRan), moduleFeed.TimesRan, ssuff(moduleFeed.TimesRan),
					moduleFeed.WaitMins, ssuff(moduleFeed.WaitMins),
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

		//#region Instagram Accounts
		"instagram-new": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
					Group:    feedInstagramAccount,
					Name:     newFeed.Name,
					Ref:      newFeed.ID,
					Config:   newFeed,
					WaitMins: waitMins,
				})
				go startFeed(&feeds[feedIndex])
			}
		},
		"instagram-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "wip"},
			})
		},
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
		"instagram-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
		},
		"instagram-show": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "wip"},
			})
		},
		//#endregion

		//#region RSS Feeds
		"rss-new": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

				// New Feed
				var newFeed configModuleRssFeed
				newFeed.Destinations = []feedDestination{{Channel: i.ChannelID}}
				if opt, ok := optionMap["url"]; ok {
					newFeed.URL = opt.StringValue()
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				// Identifiers are empty
				if newFeed.Name == "" || newFeed.URL == "" {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Config name or feed identifier was empty... Try again!"},
					})
					return
				}
				// Doesn't exist
				if existsRssConfig(newFeed.Name) {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "RSS Feed already exists with that name..."},
					})
					return
				}
				// Optional Vars
				if opt, ok := optionMap["tag"]; ok {
					tagged := opt.UserValue(s)
					if tagged != nil {
						destClone := newFeed.Destinations
						for key, destination := range destClone {
							if destination.Channel == i.ChannelID {
								newFeed.Destinations[key].Tags = []string{tagged.ID}
							}
						}
					}
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
				// Optional Vars -Lists
				if opt, ok := optionMap["blacklist"]; ok {
					var list []string
					list = append(list, strings.Split(opt.StringValue(), "|")...)
					newFeed.Blacklist = append(newFeed.Blacklist, list)
				}
				if opt, ok := optionMap["whitelist"]; ok {
					var list []string
					list = append(list, strings.Split(opt.StringValue(), "|")...)
					newFeed.Whitelist = append(newFeed.Whitelist, list)
				}
				if opt, ok := optionMap["list-type"]; ok {
					newFeed.ListType = opt.StringValue()
				}

				// Finalize
				feedIndex := len(rssConfig.Feeds)                  // cache index for new routine
				rssConfig.Feeds = append(rssConfig.Feeds, newFeed) // add new feed to config
				if err := saveModuleConfig(feedRSS); err != nil {  // save config
					log.Println(color.HiRedString("error saving config")) //TODO:
				} else { // success
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

				// Start routine for new feed
				waitMins := rssConfig.WaitMins
				if newFeed.WaitMins != nil {
					waitMins = *newFeed.WaitMins
				}
				feeds = append(feeds, moduleFeed{
					Group:    feedRSS,
					Name:     newFeed.Name,
					Ref:      "\"" + newFeed.URL + "\"",
					Config:   newFeed,
					WaitMins: waitMins,
				})
				go startFeed(&feeds[feedIndex])
			}
		},
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

				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()

					if !existsRssConfig(name) {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{Content: "No RSS Feed exists with that name..."},
						})
						return
					} else {
						config := getRssConfig(name) // point to it so it modifies source
						config.Destinations = append(config.Destinations, feedDestination{Channel: i.ChannelID})

						// Save
						updateRssConfig(config.Name, *config)
						if err := saveModuleConfig(feedRSS); err != nil {
							log.Println(color.HiRedString("error saving config")) //TODO:
						} else { // success
							reply := "Modified RSS Feed! Saved to config..."
							json, err := json.MarshalIndent(config, "", "\t")
							if err == nil {
								reply += fmt.Sprintf("\n```json\n%s```", json)
							}
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{Content: reply},
							})
						}
						// Update Live
						if !updateFeedConfig(config.Name, feedRSS, *config) {
							log.Println(color.HiRedString("failed to update feed %s/%s...", getFeedTypeName(feedRSS), config.Name))
						}
					}
				}
			}
		},
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

				if opt, ok := optionMap["name"]; !ok {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Config name identifier is empty... Try again!"},
					})
					return
				} else {
					feedName := opt.StringValue()
					if !existsRssConfig(feedName) {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{Content: "No feed config exists with that name..."},
						})
						return
					} else {
						config := getRssConfig(feedName) // point to it so it modifies source

						// Optional Vars
						if opt, ok := optionMap["change-url"]; ok {
							config.URL = opt.StringValue()
						}
						if opt, ok := optionMap["tag"]; ok {
							tagged := opt.UserValue(s)
							if tagged != nil {
								destClone := config.Destinations
								for key, destination := range destClone {
									if destination.Channel == i.ChannelID {
										config.Destinations[key].Tags = append(config.Destinations[key].Tags, tagged.ID)
									}
								}
							}
						}
						if opt, ok := optionMap["wait"]; ok {
							v := int(opt.IntValue())
							config.WaitMins = &v
						}
						if opt, ok := optionMap["avatar"]; ok {
							v := opt.StringValue()
							config.Avatar = &v
						}
						if opt, ok := optionMap["username"]; ok {
							v := opt.StringValue()
							config.Username = &v
						}
						if opt, ok := optionMap["twitter"]; ok {
							if twitterClient == nil {
								//TODO: log
								s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{Content: "Twitter Client is not connected..."},
								})
								return
							} else {
								handle := opt.StringValue()
								userResults, err := twitterClient.GetUsersLookup(handle, url.Values{})
								if err == nil {
									//TODO: log
									s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
										Type: discordgo.InteractionResponseChannelMessageWithSource,
										Data: &discordgo.InteractionResponseData{Content: "ERROR FETCHING USERS..."},
									})
									return
								} else {
									if len(userResults) > 0 {
										config.UseTwitter = &userResults[0].IdStr
									} else {
										//TODO: log
										s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
											Type: discordgo.InteractionResponseChannelMessageWithSource,
											Data: &discordgo.InteractionResponseData{Content: "No Twitter users found for this handle..."},
										})
										return
									}
								}
							}
						}
						// Optional Vars -Lists
						if opt, ok := optionMap["blacklist"]; ok {
							var list []string
							list = append(list, strings.Split(opt.StringValue(), "|")...)
							config.Blacklist = append(config.Blacklist, list)
						}
						if opt, ok := optionMap["whitelist"]; ok {
							var list []string
							list = append(list, strings.Split(opt.StringValue(), "|")...)
							config.Whitelist = append(config.Whitelist, list)
						}
						if opt, ok := optionMap["list-type"]; ok {
							config.ListType = opt.StringValue()
						}
						if opt, ok := optionMap["blacklist-url"]; ok {
							var list []string
							list = append(list, strings.Split(opt.StringValue(), "|")...)
							config.BlacklistURL = append(config.BlacklistURL, list)
						}

						// Save
						updateRssConfig(config.Name, *config)
						if err := saveModuleConfig(feedRSS); err != nil {
							log.Println(color.HiRedString("error saving config")) //TODO:
						} else { // success
							reply := "Modified RSS Feed! Saved to config..."
							json, err := json.MarshalIndent(config, "", "\t")
							if err == nil {
								reply += fmt.Sprintf("\n```json\n%s```", json)
							}
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{Content: reply},
							})
						}
						// Update Live
						if !updateFeedConfig(config.Name, feedRSS, *config) {
							log.Println(color.HiRedString("failed to update feed %s/%s...", getFeedTypeName(feedRSS), config.Name))
						}
					}
				}
			}
		},
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
						deletedConfig := deleteRssConfig(name)
						if deletedConfig {
							log.Println("deleted config")
						} else {
							log.Println("FAILED TO DELETE CONFIG")
						}
						deletedFeed := deleteFeed(name, feedRSS)
						if deletedFeed {
							log.Println("deleted feed")
						} else {
							log.Println("FAILED TO DELETE FEED")
						}

						// Save
						if err := saveModuleConfig(feedRSS); err != nil {
							log.Println(color.HiRedString("error saving config")) //TODO:
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{Content: "error saving config"},
							})
						} else { // success
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{Content: "deleted?"},
							})
						}
					}
				}
			}
		},
		"rss-show": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
						feed := getModuleFeed(name, feedRSS)

						reply := fmt.Sprintf("**RSS Feed: %s**", feed.Name)
						reply += fmt.Sprintf("\n_Ran %s, runs every %d minutes, ran %d time%s since launch_",
							humanize.Time(feed.LastRan), feed.WaitMins, feed.TimesRan, ssuff(feed.TimesRan))

						config := getRssConfig(name)
						// Append Config
						json, err := json.MarshalIndent(config, "", "\t")
						if err == nil {
							reply += fmt.Sprintf("\n```json\n%s```", json)
						}

						// Send
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{Content: reply},
						})
					}
				}
			}
		},
		//#endregion

		//#region Twitter Accounts
		"twitter-new": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
			if authorUser == nil {
				//TODO: log
				return
			}
			if !isBotAdmin(authorUser.ID) {
				//TODO: log
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: commandNotAdmin},
				})
				return
			} else {
				options := i.ApplicationCommandData().Options
				optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
				for _, opt := range options {
					optionMap[opt.Name] = opt
				}

				var newFeed configModuleTwitterAccount
				newFeed.Destinations = []string{i.ChannelID}
				if opt, ok := optionMap["handle"]; ok {
					if twitterClient == nil {
						//TODO: log
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{Content: "Twitter Client is not connected..."},
						})
						return
					} else {
						handle := opt.StringValue()
						userResults, err := twitterClient.GetUsersLookup(handle, url.Values{})
						if err == nil {
							//TODO: log
							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionResponseChannelMessageWithSource,
								Data: &discordgo.InteractionResponseData{Content: "ERROR FETCHING USERS..."},
							})
							return
						} else {
							if len(userResults) > 0 {
								userResult := userResults[0]
								newFeed.ID = userResult.IdStr
							} else {
								//TODO: log
								s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
									Type: discordgo.InteractionResponseChannelMessageWithSource,
									Data: &discordgo.InteractionResponseData{Content: "No Twitter users found for this handle..."},
								})
								return
							}
						}
					}
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				// Doesn't exist
				if existsTwitterConfig(newFeed.Name) {
					//TODO: log
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{Content: "Twitter Account already exists with that name..."},
					})
					return
				}
				if newFeed.Name == "" || newFeed.ID == "" {
					//TODO: log
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
					//TODO: log log
					log.Println(color.HiRedString("error saving config"))
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
					Group:    feedTwitterAccount,
					Name:     newFeed.Name,
					Ref:      newFeed.ID,
					Config:   newFeed,
					WaitMins: waitMins,
				})
				go startFeed(&feeds[feedIndex])
			}
		},
		"twitter-add": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "wip"},
			})
		},
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
		"twitter-show": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "wip"},
			})
		},
		//#endregion

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
