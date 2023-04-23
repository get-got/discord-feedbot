package main

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gtuk/discordwebhook"
)

func getWebhookForChannel(channel string) (*discordgo.Webhook, error) {
	webhooks, err := discord.ChannelWebhooks(channel)
	if err != nil {
		return nil, err
	}

	// Find
	for _, webhook := range webhooks {
		if webhook != nil {
			if webhook.Name == "FEEDBOT" {
				return webhook, nil
			}
		}
	}

	// Create
	newWebhook, err := discord.WebhookCreate(channel, "FEEDBOT", "")
	if err != nil {
		return nil, err
	}
	return newWebhook, nil
}

func getWebhookURL(webhook *discordgo.Webhook) string {
	if webhook != nil {
		return fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", webhook.ID, webhook.Token)
	}
	return ""
}

func sendWebhook(channel string, ref string, webhookData discordwebhook.Message, module string) error {
	webhook, err := getWebhookForChannel(channel)
	if err != nil {
		return err
	}
	webhookURL := getWebhookURL(webhook)
	if webhookURL == "" {
		return errors.New("error parsing webhook url")
	} else {
		if err = discordwebhook.SendMessage(webhookURL, webhookData); err != nil {
			return err
		} else {
			refLogSent(ref, channel, module)
		}
	}

	return nil
}
