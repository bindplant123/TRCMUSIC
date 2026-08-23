package modules

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const farewellText = "Thanks for using alone x music! Stay tuned for more music."

func createFarewellAnnouncement() (string, error) {
	if err := os.MkdirAll("downloads", 0755); err != nil {
		return "", err
	}

	query := url.Values{
		"ie":     {"UTF-8"},
		"client": {"tw-ob"},
		"tl":     {"en"},
		"q":      {farewellText},
	}
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Get("https://translate.google.com/translate_tts?" + query.Encode())
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TTS request returned %s", response.Status)
	}

	file, err := os.CreateTemp("downloads", "farewell_*.mp3")
	if err != nil {
		return "", err
	}
	path := file.Name()
	if _, err = io.Copy(file, response.Body); err != nil {
		file.Close()
		removeAnnouncement(path)
		return "", err
	}
	if err = file.Close(); err != nil {
		removeAnnouncement(path)
		return "", err
	}
	return path, nil
}

func removeAnnouncement(path string) error {
	return os.Remove(path)
}