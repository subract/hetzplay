package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/op/go-logging"
	"github.com/subract/hetzplay/handlers"
	"github.com/subract/hetzplay/hetzner"
)

// TODO: figure out the proper way to do versioning
var version string = "0.1.0"

// Bot parameters
// TODO: Migrate to env vars/conf file
var (
	GuildID         = flag.String("guild", "", "Guild ID. If not passed - bot registers commands globally")
	BotToken        = flag.String("discord_token", "", "Discord bot access token")
	HetznerToken    = flag.String("hetzner_token", "", "Hetzner API token")
	ServerName      = flag.String("server", "", "Hetzner server name. Must be unique within project.")
	BackupSnapCount = flag.Uint("backup_snaps", 1, "How many backup snapshots to keep")
	LogLevel        = flag.String("log", "notice", "One of [critical error warning notice info debug]")
)

var session *discordgo.Session
var client *hcloud.Client
var server *hcloud.Server
var log = logging.MustGetLogger("log")

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "start",
		Description: "Start the Minecraft server",
	},
}

// Define a closure to pass client/ServerName context to handler
func startCommandWithContext() func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handlers.StartCommandHandler(s, i, client, *ServerName, log)
	}
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"start": startCommandWithContext(),
}

func init() {
	// TODO: Validate args
	flag.Parse()

	// Set up logging
	err := initializeLogging(*LogLevel)
	if err != nil { // bad log level
		fmt.Println(err)
		os.Exit(1)
	}

	log.Debug("Initializing Discord bot")
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
		log.Noticef("Logged in to Discord as %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	// Create Hetzner client
	client = hcloud.NewClient(hcloud.WithToken(*HetznerToken), hcloud.WithApplication("hetzplay", version))
}

func main() {
	// Get snaps
	snaps, err := hetzner.ListSnapshots(client, *ServerName)
	if err != nil {
		log.Fatalf("Failed to list snapshots: %v", err)
	}
	for _, snap := range snaps {
		log.Info("Found snap", snap.Description, snap.Created, snap.ID)
	}

	// Check if first run
	if len(snaps) == 0 {
		log.Notice("It looks like this is your first time running Hetzplay.")

		// Verify server exists
		server, _, err = client.Server.GetByName(context.Background(), *ServerName)
		if err != nil {
			log.Fatalf("Failed to get server: %v", err)
		}
		if server == nil {
			log.Fatal("The server must be running the first time hetzplay is executed.")
		}

		log.Notice("Taking an initial snapshot of your server")
		snap, err := hetzner.TakeSnapshot(client, server, 0)
		if err != nil {
			log.Fatalf("Failed to create initial snapshot: %v", err)
		}
		log.Notice("Created snap", snap.Description)
	}

	// DEBUG: Prune snapshots
	hetzner.PruneSnapshots(client, *ServerName, *BackupSnapCount)

	log.Debug("Opening bot session.")
	err = session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Debug("Adding commands to bot session.")
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
	log.Notice("Press Ctrl+C to exit")
	<-stop

	// Remove all registered commands
	log.Notice("Shutting down.")
	log.Debug("Removing registered bot commands")

	registeredCommands, err := session.ApplicationCommands(session.State.User.ID, *GuildID)
	if err != nil {
		log.Fatalf("Could not fetch registered commands: %v", err)
	}

	for _, v := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, *GuildID, v.ID)
		if err != nil {
			log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}
