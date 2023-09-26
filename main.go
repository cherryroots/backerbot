package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	Token := os.Getenv("TOKEN")

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Panic("error creating Discord session,", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	dg.ShouldReconnectOnError = true

	err = dg.Open()
	if err != nil {
		log.Panic("error opening connection,", err)
		return
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Logged in as ", s.State.User.Username)
	})
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	registerCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, command := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", command)
		if err != nil {
			log.Printf("could not create '%s' command: %v", command.Name, err)
		}
		registerCommands[i] = cmd
		log.Printf("Created '%s' command", cmd.Name)
	}

	defer dg.Close()
	defer log.Println("Bot is shutting down.")

	log.Println("Awaiting commands...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

}
