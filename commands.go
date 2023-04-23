package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue = 1.0

	// https://github.com/bwmarrin/discordgo/blob/master/examples/slash_commands/main.go
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "xyz",
			Description: "Command for demonstrating options",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "string-option",
					Description: "String option",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "integer-option",
					Description: "Integer option",
					MinValue:    &integerOptionMinValue,
					MaxValue:    10,
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		"xyz": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			// This example stores the provided arguments in an []interface{}
			// which will be used to format the bot's response
			margs := make([]interface{}, 0, len(options))
			msgformat := "You learned how to use command options! " +
				"Take a look at the value(s) you entered:\n"

			// Get the value from the option map.
			// When the option exists, ok = true
			if option, ok := optionMap["string-option"]; ok {
				// Option values must be type asserted from interface{}.
				// Discordgo provides utility functions to make this simple.
				margs = append(margs, option.StringValue())
				msgformat += "> string-option: %s\n"
			}

			if opt, ok := optionMap["integer-option"]; ok {
				margs = append(margs, opt.IntValue())
				msgformat += "> integer-option: %d\n"
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf(
						msgformat,
						margs...,
					),
				},
			})
		},
	}
)
