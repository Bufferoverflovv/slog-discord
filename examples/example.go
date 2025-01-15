package main

import (
	"fmt"
	"log/slog"
	"time"

	slogdiscord "github.com/Bufferoverflovv/slog-discord"
)

const (
	WebhookURL string = "DISCORD WEBHOOK URL HERE"
)

func main() {
	default_example()
	custom_example()
}

func default_example() {
	// 1) Configure how you want your Discord logs to look
	cfg := slogdiscord.DiscordWebhookConfig{
		WebhookURL: WebhookURL,
	}

	// 2) Create your Discord handler
	discordHandler := slogdiscord.NewDiscordHandler(cfg)

	// 3) Create a slog.Logger
	logger := slog.New(discordHandler)

	// 4) Log away!
	logger.Info("Hello from Info", "version", "1.2.3")
	logger.Warn("Watch out, memory usage is high", "usage", "85%")
	logger.Error("Oh no, database connection failed", "error", "timeout")
}

func custom_example() {
	// 1) Configure how you want your Discord logs to look
	cfg := slogdiscord.DiscordWebhookConfig{
		MinLevel:   slog.LevelDebug,
		Timeout:    5 * time.Second,
		WebhookURL: WebhookURL,
		Username:   "Slog Notifications",
		LevelColors: slogdiscord.LevelColors{
			"DEBUG": 0x95a5a6, // Gray
			"INFO":  0x3498db, // Blue
			"WARN":  0xf1c40f, // Yellow
			"ERROR": 0xe74c3c, // Red
		},
		// Optional custom embed function. If omitted, a default is used.
		CustomEmbed: func(r slog.Record, lc slogdiscord.LevelColors) *slogdiscord.DiscordEmbed {
			// custom formatting
			return &slogdiscord.DiscordEmbed{
				Title:       fmt.Sprintf("Custom title: %s", r.Level.String()),
				Description: fmt.Sprintf("Custom description: %s", r.Message),
				Color:       lc[r.Level.String()],
				Timestamp:   time.Now().Format(time.RFC3339),
				Fields: []slogdiscord.EmbedField{
					{
						Name:   "Custom Field",
						Value:  "Custom Value",
						Inline: true,
					},
				},
				Footer: &slogdiscord.EmbedFooter{
					Text: "Custom Footer",
				},
			}
		},
	}

	// 3) Create your Discord handler
	discordHandler := slogdiscord.NewDiscordHandler(cfg)

	// 4) Create a slog.Logger
	logger := slog.New(discordHandler)

	// 5) Log away!
	logger.Info("Hello from Info", "version", "1.2.3")
	logger.Warn("Watch out, memory usage is high", "usage", "85%")
	logger.Error("Oh no, database connection failed", "error", "timeout")
}
