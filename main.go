package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/subract/hetzplay/hetzner"
)

// Bot parameters
var (
	GuildID      = flag.String("guild", "", "Guild ID. If not passed - bot registers commands globally")
	BotToken     = flag.String("discord_token", "", "Discord bot access token")
	HetznerToken = flag.String("hetzner_token", "", "Hetzner API token")
	ServerID     = flag.Int64("server_id", 0, "Hetzner server ID")
)

var session *discordgo.Session
var client *hcloud.Client

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "start",
			Description: "Start the Minecraft server",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// Reply to the user
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "I'll see what I can do.",
				},
			})

			// List snapshots
			hetzner.ListSnapshots(client)
		},
	}
)

func init() {
	flag.Parse()

	// Authenticate the bot
	var err error
	session, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	// Add command handlers
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Add ready handler
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Create Hetzner client
	client = hcloud.NewClient(hcloud.WithToken(*HetznerToken))
}

func main() {

	// Check if first run
	if hetzner.ListSnapshots(client) == nil {
		fmt.Println("It looks like this is your first time using hetzplay on this server.")
	}

	// Connect the bot session
	err := session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Add commands to bot session
	log.Println("Adding commands...")
	for _, v := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, *GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

	// Ensure the bot session is closed when done
	defer session.Close()

	// Catch signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)    // stop on ctrl-C
	signal.Notify(stop, syscall.SIGTERM) // stop on docker stop
	log.Println("Press Ctrl+C to exit")
	<-stop

	// Remove all registered commands
	log.Println("Removing commands...")

	registeredCommands, err := session.ApplicationCommands(session.State.User.ID, *GuildID)
	if err != nil {
		log.Fatalf("Could not fetch registered commands: %v", err)
	}

	for _, v := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, *GuildID, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Println("Gracefully shutting down.")
}
