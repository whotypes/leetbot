package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/whotypes/leetbot/internal/config"
	"github.com/whotypes/leetbot/internal/data"
	"github.com/whotypes/leetbot/internal/discord"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	fmt.Printf("Starting leetbot with prefix '%s'...\n", cfg.BotPrefix)
	fmt.Println("Loading problems data...")
	problemsData, err := data.LoadAllProblems()
	if err != nil {
		log.Fatalf("Failed to load problems data: %v", err)
	}

	fmt.Printf("Loaded data for %d companies\n", len(problemsData.GetAvailableCompanies()))

    handler := discord.NewHandler(problemsData, cfg.BotPrefix)

    // Wire up process storage if configured
    if cfg.FirestoreProjectID != "" {
        ctx := context.Background()
        fs, err := data.NewFirestoreClient(ctx, cfg.FirestoreProjectID, cfg.FirestoreDatabaseID)
        if err != nil {
            log.Printf("Warning: failed to initialize Firestore storage: %v", err)
        } else {
            handler.SetProcessStorage(fs)
            log.Println("✓ Process storage configured (Firestore)")
        }
    } else {
        log.Println("Process storage not configured; !process command will be disabled")
    }
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	// create a channel to signal reconnection
	reconnectChan := make(chan bool)
	handler.SetReconnectChannel(reconnectChan)

	dg.AddHandler(handler.HandleMessage)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			// let the paginator handle button clicks
			discord.PaginatorManager.OnInteractionCreate(s, i)
		case discordgo.InteractionApplicationCommand:
			handler.HandleSlashCommand(s, i)
		case discordgo.InteractionApplicationCommandAutocomplete:
			discord.HandleAutocomplete(s, i, problemsData)
		}
	})

	// add reconnection handler
	dg.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
		// check if this is a reconnection
		select {
		case <-reconnectChan:
			log.Println("🔄 Bot reconnected to Discord successfully")
			return
		default:
		}

		log.Printf("Leetbot logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		commands := discord.GetSlashCommands(problemsData)

		log.Println("Clearing old slash commands...")
		// Get all currently registered commands
		registeredCommands, err := s.ApplicationCommands(s.State.User.ID, "")
		if err != nil {
			log.Printf("Warning: failed to get registered commands: %v", err)
		} else {
			// Delete commands that are no longer in our code
			for _, registeredCmd := range registeredCommands {
				// Check if this command should still exist
				shouldExist := false
				for _, currentCmd := range commands {
					if registeredCmd.Name == currentCmd.Name {
						shouldExist = true
						break
					}
				}

				if !shouldExist {
					log.Printf("Removing old command: /%v", registeredCmd.Name)
					err := s.ApplicationCommandDelete(s.State.User.ID, "", registeredCmd.ID)
					if err != nil {
						log.Printf("Failed to delete command '%v': %v", registeredCmd.Name, err)
					}
				}
			}
		}

		log.Println("Registering slash commands...")
		for _, v := range commands {
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
			if err != nil {
				log.Printf("Cannot create '%v' command: %v", v.Name, err)
			} else {
				log.Printf("✓ Registered command: /%v", v.Name)
			}
		}
		log.Printf("Finished registering slash commands")
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord connection: %v", err)
	}
	defer dg.Close()

	// start a goroutine to handle reconnection signals
	go func() {
		for {
			select {
			case <-reconnectChan:
				log.Println("🔄 Received restart signal, restarting Discord session...")

				// close current session
				err := dg.Close()
				if err != nil {
					log.Printf("Error closing Discord session: %v", err)
				}

				// wait a bit before reconnecting
				time.Sleep(1 * time.Second)

				// reopen the session
				err = dg.Open()
				if err != nil {
					log.Printf("Error reopening Discord session: %v", err)
					// if reconnection fails, we can't really do much more from here
					// the main process will need to be restarted
				} else {
					log.Println("✅ Bot restarted successfully")
				}
			}
		}
	}()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Shutting down bot...")
}
