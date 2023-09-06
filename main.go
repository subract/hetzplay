package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"github.com/subract/hetzplay/hetzner"
)

// TODO: figure out the proper way to do versioning
var version string = "0.1.0"
var serverMgr *hetzner.ServerManager
var session *discordgo.Session
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
		// handlers.StartCommandHandler(s, i, client, *ServerName, log)
	}
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"start": startCommandWithContext(),
}

func init() {

	// Handle configuration
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetEnvPrefix("hetzplay")
	viper.BindEnv("log_level")

	viper.SetDefault("backup_snaps", 1)
	viper.SetDefault("log_level", "notice")

	err := viper.ReadInConfig()
	if err != nil { // config
		log.Error(err)
		os.Exit(1)
	}

	// Set up logging
	err = initializeLogging(viper.GetString("log_level"))
	if err != nil { // bad log level
		fmt.Println(err)
		os.Exit(1)
	}

	log.Notice("Starting Hetzplay", version)

	log.Debug("Initializing Discord bot")

	session, err = discordgo.New("Bot " + viper.GetString(("discord.token")))
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
		log.Noticef("Logged in to Discord as %v#%v.", s.State.User.Username, s.State.User.Discriminator)
	})

	// Create server manager
	serverMgr, err = hetzner.NewServerManager(viper.GetString("hetzner.server.name"),
		viper.GetInt("hetzner.server.backup_snaps"),
		viper.GetString("hetzner.token"),
		version,
		log)

	if err != nil {
		log.Fatalf("failed to initialize server manager: %s", err)
	}
	log.Notice("Logged in to Hetzner")

}

func main() {
	log.Debug("Opening bot session.")
	err := session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Debug("Adding commands to bot session.")
	for _, v := range commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, viper.GetString("discord.guild_id"), v)
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

	registeredCommands, err := session.ApplicationCommands(session.State.User.ID, viper.GetString("discord.guild_id"))
	if err != nil {
		log.Fatalf("Could not fetch registered commands: %v", err)
	}

	for _, v := range registeredCommands {
		err := session.ApplicationCommandDelete(session.State.User.ID, viper.GetString("discord.guild_id"), v.ID)
		if err != nil {
			log.Fatalf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}
