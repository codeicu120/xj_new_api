package ucp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TelegramNotifier interface {
	Send(ctx context.Context, botAPIKey string, chatID string, text string) error
}

type httpTelegramNotifier struct {
	client *http.Client
}

func (n httpTelegramNotifier) Send(ctx context.Context, botAPIKey string, chatID string, text string) error {
	botAPIKey = strings.TrimSpace(botAPIKey)
	chatID = strings.TrimSpace(chatID)
	if botAPIKey == "" || chatID == "" || strings.TrimSpace(text) == "" {
		return nil
	}
	client := n.client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	endpoint := "https://api.telegram.org/bot" + url.PathEscape(botAPIKey) + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(url.Values{
		"chat_id": {chatID},
		"text":    {text},
	}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram status %d", resp.StatusCode)
	}
	return nil
}
