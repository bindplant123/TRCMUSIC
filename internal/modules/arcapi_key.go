package modules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/locales"
)

func init() {
	helpTexts["/setarcapi"] = `<i>Update the Arc Music API key and restart the bot.</i>

<u>Usage:</u>
<code>/setarcapi YOUR_ARC_API_KEY</code>

<b>🔒 Restrictions:</b> Owner only.`
}

func handleSetArcAPI(m *tg.NewMessage) error {
	apiKey := strings.TrimSpace(m.Args())
	if apiKey == "" || strings.ContainsAny(apiKey, " \t\r\n") {
		m.Reply(F(m.ChannelID(), "setarcapi_usage"))
		return tg.ErrEndGroup
	}

	if config.RenderAPIKey != "" || config.RenderServiceID != "" {
		if config.RenderAPIKey == "" || config.RenderServiceID == "" {
			m.Reply(F(m.ChannelID(), "setarcapi_render_config"))
			return tg.ErrEndGroup
		}
		if err := updateRenderArcAPIKey(apiKey); err != nil {
			m.Reply(F(m.ChannelID(), "setarcapi_failed", locales.Arg{"error": err.Error()}))
			return tg.ErrEndGroup
		}
	} else if err := updateLocalArcAPIKey(apiKey); err != nil {
		m.Reply(F(m.ChannelID(), "setarcapi_failed", locales.Arg{"error": err.Error()}))
		return tg.ErrEndGroup
	}

	config.ArcAPIKey = apiKey
	_ = os.Setenv("ARC_API_KEY", apiKey)
	m.Reply(F(m.ChannelID(), "setarcapi_success"))
	return handleRestart(m)
}

func updateLocalArcAPIKey(apiKey string) error {
	path := ".env"
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	found := false
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "ARC_API_KEY=") {
			lines[index] = "ARC_API_KEY=" + apiKey
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, "ARC_API_KEY="+apiKey)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0600)
}

func updateRenderArcAPIKey(apiKey string) error {
	payload, err := json.Marshal(map[string]string{"value": apiKey})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	endpoint := "https://api.render.com/v1/services/" + url.PathEscape(config.RenderServiceID) + "/env-vars/ARC_API_KEY"
	request, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+config.RenderAPIKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("Render API returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	deployEndpoint := "https://api.render.com/v1/services/" + url.PathEscape(config.RenderServiceID) + "/deploys"
	deployRequest, err := http.NewRequest(http.MethodPost, deployEndpoint, nil)
	if err != nil {
		return err
	}
	deployRequest.Header.Set("Authorization", "Bearer "+config.RenderAPIKey)
	deployResponse, err := client.Do(deployRequest)
	if err != nil {
		return err
	}
	defer deployResponse.Body.Close()
	if deployResponse.StatusCode < http.StatusOK || deployResponse.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Render deploy API returned %s", deployResponse.Status)
	}

	return nil
}