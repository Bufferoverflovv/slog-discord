package slogdiscord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// CustomEmbed allows the user to provide a function that converts a slog.Record
// into a DiscordEmbed, allowing for custom titles, descriptions, footers, etc.
type CustomEmbed func(r slog.Record, levelColors LevelColors) *DiscordEmbed

// DiscordWebhookConfig configures how logs are sent to Discord.
type DiscordWebhookConfig struct {
	MinLevel    slog.Level    // Set the minimum slog level (Default: Debug) : Optional
	Timeout     time.Duration // Set the timeout (Default: 5 Seconds) : Optional
	WebhookURL  string        // The webhook URL from discord
	Username    string        // Set a custom username : Optional
	AvatarURL   string        // Set a custom avatar : Optional
	LevelColors LevelColors   // Customise the colours for each slog level : Optional
	CustomEmbed CustomEmbed   // Customise the embed content : Optional
}

// DiscordHandler is our slog.Handler implementation that sends logs directly to Discord.
type DiscordHandler struct {
	config DiscordWebhookConfig
}

// NewDiscordHandler constructs a slog.Handler that sends logs to Discord, with **no** next handler.
func NewDiscordHandler(cfg DiscordWebhookConfig) slog.Handler {
	return &DiscordHandler{
		config: cfg,
	}
}

// Enabled can filter logs by level if you want. For simplicity, return true for all levels.
func (h *DiscordHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	if h.config.MinLevel == 0 {
		h.config.MinLevel = slog.LevelDebug
	}
	return lvl >= h.config.MinLevel
}

// Handle is called for every log record; we build the Discord embed and send it.
func (h *DiscordHandler) Handle(ctx context.Context, r slog.Record) error {
	// Build a Discord embed. If user provided a CustomEmbed, use it; otherwise default
	var embed *DiscordEmbed
	if h.config.CustomEmbed != nil {
		embed = h.config.CustomEmbed(r, h.config.LevelColors)
	} else {
		embed = DefaultEmbed(r, h.config.LevelColors)
	}

	// Create the full payload
	payload := Payload{
		Username:  h.config.Username,
		AvatarURL: h.config.AvatarURL,
		Embeds:    []DiscordEmbed{*embed},
	}

	// Send to Discord
	if err := h.sendToDiscord(payload, h.config.Timeout); err != nil {
		return fmt.Errorf("failed to send log to Discord: %w", err)
	}

	return nil
}

// WithAttrs is called if user adds attributes to the logger globally. You can store them
// if you want, but if you do nothing, the attributes will just be ignored at the handler level.
func (h *DiscordHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// No-op or you could clone "h" and store them
	return h
}

// WithGroup is called when the user groups attributes under a key. Same logic: no-op, or store them.
func (h *DiscordHandler) WithGroup(_ string) slog.Handler {
	// No-op
	return h
}

// DiscordLogPayload is the top-level JSON structure for sending embed(s) to Discord.
type Payload struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds"`
}

func (h *DiscordHandler) sendToDiscord(payload Payload, timeOut time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", h.config.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if h.config.Timeout == 0 {
		timeOut = 5 * time.Second
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeOut}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
