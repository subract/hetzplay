package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/subract/hetzplay/handlers"
	"github.com/subract/hetzplay/hetzner"
)

// TODO: figure out the proper way to do versioning
var version string = "0.1.0"

// Bot parameters
// TODO: Migrate to env vars/conf file
var (
	GuildID      = flag.String("guild", "", "Guild ID. If not passed - bot registers commands globally")
	BotToken     = flag.String("discord_token", "", "Discord bot access token")
	HetznerToken = flag.String("hetzner_token", "", "Hetzner API token")
	ServerName   = flag.String("server", "", "Hetzner server name")
)

var session *discordgo.Session
var client *hcloud.Client
var server *hcloud.Server

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "start",
			Description: "Start the Minecraft server",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// Need to pass the client and server name to the handler
		// ChatGPT taught me to use a closure to do this, hope that's appropriate
		"start": func() func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
				handlers.StartCommandHandler(s, i, client, *ServerName)
			}
		}(),
	}
)

func init() {
	// TODO: Validate args
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
	client = hcloud.NewClient(hcloud.WithToken(*HetznerToken), hcloud.WithApplication("hetzplay", version))
}

func main() {

	snaps, err := hetzner.ListSnapshots(client, *ServerName)
	if err != nil {
		log.Fatalf("Failed to list snapshots: %v", err)
	}

	// Check if first run
	if len(snaps) == 0 {
		fmt.Println("It looks like this is your first time running Hetzplay.")

		// Verify server exists
		server, _, err = client.Server.GetByName(context.Background(), *ServerName)
		if err != nil {
			log.Fatalf("Failed to get server: %v", err)
		}
		if server == nil {
			log.Fatal("The server must be running the first time hetzplay is executed.")
		}

		fmt.Print("Taking an initial snapshot of your server... ")
		_, err := hetzner.TakeSnapshot(client, server, 0)
		if err != nil {
			log.Fatalf("Failed to create initial snapshot: %v", err)
		}
		fmt.Println("done.")
	}

	// Connect the bot session
	err = session.Open()
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
