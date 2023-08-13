package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/op/go-logging"
)

func StartCommandHandler(s *discordgo.Session,
	i *discordgo.InteractionCreate,
	client *hcloud.Client,
	serverName string,
	log *logging.Logger) {
	// Reply to the user
	log.Infof("Got start command")
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "I'll see what I can do.",
		},
	})
}
