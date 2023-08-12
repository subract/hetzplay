package handlers

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/subract/hetzplay/hetzner"
)

func StartCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate, client *hcloud.Client, serverName string) {
	// Reply to the user
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "I'll see what I can do.",
		},
	})

	// List snapshots
	snapshots, err := hetzner.ListSnapshots(client, serverName)
	if err != nil {
		log.Fatalf("error retrieving snapshots: %s\n", err)
	}

	for _, snapshot := range snapshots {
		fmt.Printf("Found snapshot %s", snapshot.Description)
	}
	return
}
