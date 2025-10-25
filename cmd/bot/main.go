package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
	dg.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
		log.Printf("Leetbot logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		commands := discord.GetSlashCommands(problemsData)

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

	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord connection: %v", err)
	}
	defer dg.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Shutting down bot...")
}
