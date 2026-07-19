package bark

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Client sends notifications to a Bark-compatible server.
type Client struct {
	ServerURL string
	DeviceKey string
	HTTP      *http.Client
}

func NewClient(serverURL, deviceKey string) Client {
	return Client{ServerURL: strings.TrimRight(strings.TrimSpace(serverURL), "/"), DeviceKey: deviceKey, HTTP: &http.Client{}}
}

func (c Client) Send(ctx context.Context, title, body string) error {
	if strings.TrimSpace(c.ServerURL) == "" || strings.TrimSpace(c.DeviceKey) == "" {
		return errors.New("Bark server URL and device key are required")
	}
	base, err := url.Parse(c.ServerURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return errors.New("Bark server URL is invalid")
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/" + url.PathEscape(c.DeviceKey) + "/" + url.PathEscape(title) + "/" + url.PathEscape(body)
	query := base.Query()
	query.Set("level", "critical")
	query.Set("volume", "5")
	query.Set("call", "1")
	query.Set("sound", "minuet")
	query.Set("group", "codex")
	base.RawQuery = query.Encode()

	client := c.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return errors.New("could not build Bark request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Bark request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Bark request returned HTTP %d", resp.StatusCode)
	}
	return nil
}
