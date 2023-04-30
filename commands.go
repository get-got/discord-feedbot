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

var (
	nameCommandOpt = []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "Unique Feed Name",
			Required:    true,
		},
	}

	genericCommandOpts = []*discordgo.ApplicationCommandOption{
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
			Name:        "color",
			Description: "Webhook Embed Color",
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

	twitterOpts = []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "change-handle",
			Description: "Change Twitter Handle (@)",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "exclude-replies",
			Description: "Exclude Replies",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "include-retweets",
			Description: "Include Retweets",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "filter-type",
			Description: "Filter Type",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "ALL",
					Value: "all",
				},
				{
					Name:  "Media Only",
					Value: "media",
				},
				{
					Name:  "Text Only",
					Value: "text",
				},
				{
					Name:  "Images Only",
					Value: "image",
				},
				{
					Name:  "Videos Only",
					Value: "video",
				},
				{
					Name:  "Links Only",
					Value: "link",
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
			Description: "List active feeds; filterable",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "filter",
					Description: "Filter by group",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "ALL (Default)",
							Value: -1,
						},
						{
							Name:  "Instagram Accounts",
							Value: feedInstagramAccount,
						},
						{
							Name:  "RSS Feeds",
							Value: feedRSS,
						},
						{
							Name:  "Twitter Accounts",
							Value: feedTwitterAccount,
						},
					},
				},
			},
		},
		//#endregion

		//#region RSS Feeds
		{
			Name:        "rss-new",
			Description: "Add a new feed",
			Options: append(append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "RSS Feed URL",
				Required:    true,
			}}, genericCommandOpts...), []*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "twitter",
				Description: "Twitter for Username & Avatar",
				Required:    false,
			}}...),
		},
		{
			Name:        "rss-add",
			Description: "Add this channel to an existing feed",
			Options:     nameCommandOpt,
		},
		{
			Name:        "rss-modify",
			Description: "Modify an existing feed",
			Options: append(genericCommandOpts, []*discordgo.ApplicationCommandOption{
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
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "twitter",
					Description: "Twitter for Username & Avatar",
					Required:    false,
				},
			}...),
		},
		{
			Name:        "rss-delete",
			Description: "Delete an existing feed",
			Options:     nameCommandOpt,
		},
		{
			Name:        "rss-show",
			Description: "Display info for an existing feed",
			Options:     nameCommandOpt,
		},
		//#endregion

		//#region Twitter Accounts
		{
			Name:        "twitter-new",
			Description: "Add a new feed",
			Options: append(append([]*discordgo.ApplicationCommandOption{{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "handle",
				Description: "Twitter Handle (@)",
				Required:    true,
			}}, genericCommandOpts...), twitterOpts...),
		},
		{
			Name:        "twitter-add",
			Description: "Add this channel to an existing feed",
			Options:     nameCommandOpt,
		},
		{
			Name:        "twitter-modify",
			Description: "Modify an existing feed",
			Options:     append(genericCommandOpts, twitterOpts...),
		},
		{
			Name:        "twitter-delete",
			Description: "Delete an existing feed",
			Options:     nameCommandOpt,
		},
		{
			Name:        "twitter-show",
			Description: "Display info for an existing feed",
			Options:     nameCommandOpt,
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

				filter := -1

				if opt, ok := optionMap["filter"]; ok {
					filter = int(opt.IntValue())
				}

				output := ""
				for _, feedThread := range feeds {
					if filter != -1 {
						if feedThread.Group != filter {
							continue
						}
					}
					output += fmt.Sprintf("\n• %s: `%s` \t\t_Last ran %s < %d time%s, every %d minute%s >_",
						getFeedTypeName(feedThread.Group), feedThread.Name,
						humanize.Time(feedThread.LastRan), feedThread.TimesRan, ssuff(feedThread.TimesRan),
						feedThread.WaitMins, ssuff(feedThread.WaitMins),
					)
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: output,
					},
				})
			}
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

				feeds = append(feeds, feedThread{
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
			//TODO: everything
			InteractionRespond("wip", s, i)
		},
		"instagram-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
			InteractionRespond("wip", s, i)
		},
		"instagram-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
			InteractionRespond("wip", s, i)
		},
		"instagram-show": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			//TODO: everything
			InteractionRespond("wip", s, i)
		},
		//#endregion

		//#region RSS Feeds
		"rss-new": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
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
					InteractionRespond("Config name or feed identifier was empty... Try again!", s, i)
					return
				}
				// Doesn't exist
				if existsRssConfig(newFeed.Name) {
					InteractionRespond("RSS Feed already exists with that name...", s, i)
					return
				}

				// Handle Options
				handleRssCmdOpts(&newFeed, optionMap, s, i)

				// Finalize
				rssConfig.Feeds = append(rssConfig.Feeds, newFeed) // add new feed to config
				if err := saveModuleConfigReply(feedRSS, newFeed, "Added new RSS feed! Saved to config...", s, i); err != nil {
					log.Println(color.HiRedString("failed to save config for %s...", getFeedTypeName(feedRSS)))
				}

				// Start new feed
				waitMins := rssConfig.WaitMins
				if newFeed.WaitMins != nil {
					waitMins = *newFeed.WaitMins
				}
				feedIndex := len(feeds)
				feeds = append(feeds, feedThread{
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
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()

					if !existsRssConfig(name) {
						InteractionRespond("No RSS Feed exists with that name...", s, i)
						return
					} else {
						config := getRssConfig(name) // point to it so it modifies source
						config.Destinations = append(config.Destinations, feedDestination{Channel: i.ChannelID})

						// Save
						updateRssConfig(config.Name, *config)
						if err := saveModuleConfigReply(feedRSS, *config, "Modified RSS Feed! Saved to config...", s, i); err != nil {
							log.Println(color.HiRedString("failed to save config for %s...", getFeedTypeName(feedRSS)))
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
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; !ok {
					InteractionRespond("Config name identifier is empty... Try again!", s, i)
					return
				} else {
					feedName := opt.StringValue()
					if !existsRssConfig(feedName) {
						InteractionRespond("No feed config exists with that name...", s, i)
						return
					} else {
						config := getRssConfig(feedName) // point to it so it modifies source

						// Handle Options
						handleRssCmdOpts(config, optionMap, s, i)

						// Save
						updateRssConfig(config.Name, *config)
						if err := saveModuleConfigReply(feedRSS, *config, "Modified RSS Feed! Saved to config...", s, i); err != nil {
							log.Println(color.HiRedString("failed to save config for %s...", getFeedTypeName(feedRSS)))
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
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()
					if !existsRssConfig(name) {
						InteractionRespond("No RSS Feed exists with that name...", s, i)
						return
					} else {
						if err := deleteRssConfig(name); err != nil {
							InteractionRespond("Error deleting feed: "+err.Error(), s, i)
							return
						}
						// Save
						if err := saveModuleConfig(feedRSS); err != nil {
							InteractionRespond("Error saving RSS config: "+err.Error(), s, i)
						} else {
							InteractionRespond("Successfully deleted feed!", s, i)
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
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()
					if !existsRssConfig(name) {
						InteractionRespond("No RSS Feed exists with that name...", s, i)
						return
					} else {
						feed := getModuleFeed(name, feedRSS)
						reply := fmt.Sprintf("**RSS Feed: %s**", feed.Name)
						reply += fmt.Sprintf("\n_Ran %s, runs every %d minutes, ran %d time%s since launch_",
							humanize.Time(feed.LastRan), feed.WaitMins, feed.TimesRan, ssuff(feed.TimesRan))
						config := getRssConfig(name)
						if err := replyConfig(*config, reply, s, i); err != nil {
							log.Println(color.HiRedString("Error replying: %s", err.Error()))
						}
						// Send
						InteractionRespond(reply, s, i)
					}
				}
			}
		},
		//#endregion

		//#region Twitter Accounts
		"twitter-new": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				// New Feed
				var newFeed configModuleTwitterAcc
				newFeed.Destinations = []feedDestination{{Channel: i.ChannelID}}
				if opt, ok := optionMap["handle"]; ok {
					if twitterClient == nil {
						//TODO: log
						InteractionRespond("Twitter Client is not connected...", s, i)
						return
					} else {
						handle := opt.StringValue()
						userResults, err := twitterClient.GetUsersLookup(handle, url.Values{})
						if err == nil {
							//TODO: log
							InteractionRespond("ERROR FETCHING USERS... "+err.Error(), s, i)
							return
						} else {
							if len(userResults) > 0 {
								userResult := userResults[0]
								newFeed.ID = userResult.IdStr
							} else {
								//TODO: log
								InteractionRespond("No Twitter users found for this handle...", s, i)
								return
							}
						}
					}
				}
				if opt, ok := optionMap["name"]; ok {
					newFeed.Name = opt.StringValue()
				}
				// Identifiers are empty
				if newFeed.Name == "" || newFeed.ID == "" {
					InteractionRespond("Config name or feed identifier was empty... Try again!", s, i)
					return
				}
				// Doesn't exist
				if existsTwitterAccConfig(newFeed.Name) {
					InteractionRespond("Twitter Account already exists with that name...", s, i)
					return
				}

				// Handle Options
				handleTwitterAccCmdOpts(&newFeed, optionMap, s, i)

				// Finalize
				twitterConfig.Accounts = append(twitterConfig.Accounts, newFeed) // add new feed to config
				if err := saveModuleConfigReply(feedTwitterAccount, newFeed, "Added new Twitter Account! Saved to config...", s, i); err != nil {
					log.Println(color.HiRedString("failed to save config for %s...", getFeedTypeName(feedTwitterAccount)))
				}

				// Start new feed
				waitMins := twitterConfig.WaitMins
				if newFeed.WaitMins != nil {
					waitMins = *newFeed.WaitMins
				}
				feedIndex := len(feeds)
				feeds = append(feeds, feedThread{
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
			authorUser := getAuthor(i)
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()

					if !existsTwitterAccConfig(name) {
						InteractionRespond("No Twitter Account exists with that name...", s, i)
						return
					} else {
						config := getTwitterAccConfig(name) // point to it so it modifies source
						config.Destinations = append(config.Destinations, feedDestination{Channel: i.ChannelID})

						// Save
						updateTwitterAccConfig(config.Name, *config)
						if err := saveModuleConfigReply(feedTwitterAccount, *config, "Modified Twitter Account! Saved to config...", s, i); err != nil {
							log.Println(color.HiRedString("failed to save config for %s...", getFeedTypeName(feedTwitterAccount)))
						}
						// Update Live
						if !updateFeedConfig(config.Name, feedTwitterAccount, *config) {
							log.Println(color.HiRedString("failed to update feed %s/%s...", getFeedTypeName(feedTwitterAccount), config.Name))
						}
					}
				}
			}
		},
		"twitter-modify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; !ok {
					InteractionRespond("Config name identifier is empty... Try again!", s, i)
					return
				} else {
					feedName := opt.StringValue()
					if !existsTwitterAccConfig(feedName) {
						InteractionRespond("No feed config exists with that name...", s, i)
						return
					} else {
						config := getTwitterAccConfig(feedName) // point to it so it modifies source

						// Handle Options
						handleTwitterAccCmdOpts(config, optionMap, s, i)

						// Save
						updateTwitterAccConfig(config.Name, *config)
						if err := saveModuleConfigReply(feedTwitterAccount, *config, "Modified Twitter Account! Saved to config...", s, i); err != nil {
							log.Println(color.HiRedString("failed to save config for %s...", getFeedTypeName(feedTwitterAccount)))
						}
						// Update Live
						if !updateFeedConfig(config.Name, feedTwitterAccount, *config) {
							log.Println(color.HiRedString("failed to update feed %s/%s...", getFeedTypeName(feedTwitterAccount), config.Name))
						}
					}
				}
			}
		},
		"twitter-delete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()
					if !existsRssConfig(name) {
						InteractionRespond("No Twitter Account exists with that name...", s, i)
						return
					} else {
						if err := deleteTwitterAccConfig(name); err != nil {
							InteractionRespond("Error deleting feed: "+err.Error(), s, i)
							return
						}
						// Save
						if err := saveModuleConfig(feedTwitterAccount); err != nil {
							InteractionRespond("Error saving Twitter Account config: "+err.Error(), s, i)
						} else {
							InteractionRespond("Successfully deleted feed!", s, i)
						}
					}
				}
			}
		},
		"twitter-show": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			authorUser := getAuthor(i)
			if authorUser == nil {
				return
			}
			if !isBotAdmin(authorUser.ID) {
				InteractionRespond(commandNotAdmin, s, i)
			} else {
				optionMap := interactionOptMap(i)
				if opt, ok := optionMap["name"]; ok {
					name := opt.StringValue()
					if !existsRssConfig(name) {
						InteractionRespond("No Twitter Account exists with that name...", s, i)
						return
					} else {
						feed := getModuleFeed(name, feedTwitterAccount)
						reply := fmt.Sprintf("**Twitter Account: %s**", feed.Name)
						reply += fmt.Sprintf("\n_Ran %s, runs every %d minutes, ran %d time%s since launch_",
							humanize.Time(feed.LastRan), feed.WaitMins, feed.TimesRan, ssuff(feed.TimesRan))
						config := getTwitterAccConfig(name)
						if err := replyConfig(*config, reply, s, i); err != nil {
							log.Println(color.HiRedString("Error replying: %s", err.Error()))
						}
						// Send
						InteractionRespond(reply, s, i)
					}
				}
			}
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

func interactionOptMap(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
}

func InteractionRespond(content string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
	return nil
}
