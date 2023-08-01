package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Variables used for command line parameters
var (
	discordToken string
	hetznerToken string
)

func init() {

	flag.StringVar(&discordToken, "d", "", "Discord bot token")
	flag.StringVar(&hetznerToken, "a", "", "Hetzner API token")
	flag.Parse()
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Receive events for incoming guild messages and DMs
	dg.Identify.Intents = discordgo.IntentsGuildMessages + discordgo.IntentDirectMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// Handle incoming messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	client := hcloud.NewClient(hcloud.WithToken(hetznerToken))

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "Yo bot! Put a round into dat server!" || m.Content == "off" {
		s.ChannelMessageSend(m.ChannelID, "You got it, boss.")
		stopServer(client)
		s.ChannelMessageSend(m.ChannelID, "Server's sleeping with the fishes.")
	}

	if m.Content == "Wait! Not _that_ server!" || m.Content == "on" {
		s.ChannelMessageSend(m.ChannelID, "Oh shit, gimme a sec")
		startServer(client)
		s.ChannelMessageSend(m.ChannelID, "phew, close one")
	}
}

// Stops a server
func stopServer(client *hcloud.Client) {
	server, _, err := client.Server.GetByName(context.Background(), "games1")
	if err != nil {
		log.Fatalf("error retrieving server: %s\n", err)
	}
	if server == nil {
		fmt.Println("Server not found")
	}

	_, _, err = client.Server.Shutdown(context.Background(), server)

	if err != nil {
		log.Fatalf("error shutting down server: %s\n", err)
	}

	// Wait for server to complete shutdown
	waitForServerStatus(client, "off")

}

// Starts a server
func startServer(client *hcloud.Client) {
	server, _, err := client.Server.GetByName(context.Background(), "games1")
	if err != nil {
		log.Fatalf("error retrieving server: %s\n", err)
	}
	if server == nil {
		fmt.Println("Server not found")
	}

	_, _, err = client.Server.Poweron(context.Background(), server)

	if err != nil {
		log.Fatalf("error powering on server: %s\n", err)
	}

	waitForServerStatus(client, "running")
}

// waitForServerStatus waits until a server has a particular status
// It checks every two seconds, with a built-in timeout of one minute
func waitForServerStatus(client *hcloud.Client, targetStatus string) {
	for i := 0; i < 30; i++ {
		time.Sleep(time.Duration(2 * time.Second))
		server, _, _ := client.Server.GetByName(context.Background(), "games1")

		fmt.Println("Server is", server.Status)
		if string(server.Status) == targetStatus {
			break
		}
	}
}
