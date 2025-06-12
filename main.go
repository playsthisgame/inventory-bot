package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func checkInventory(url string) (bool, string, error) {
	var title string
	var available bool

	c := colly.NewCollector(
		colly.UserAgent(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/122.0.0.0 Safari/537.36",
		),
	)

	c.OnHTML("title", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)
	})

	c.OnHTML(`button[data-test-id="add-to-cart"]`, func(e *colly.HTMLElement) {
		if strings.Contains(strings.ToLower(e.Text), "add to cart") {
			available = true
		}
	})

	err := c.Visit(url)
	return available, title, err
}

func sendDiscordAlert(content string) error {
	webhookURL := "WEB HOOK HERE"
	if webhookURL == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL not set")
	}

	// Use a struct to safely build the JSON
	payload := map[string]string{
		"content": content,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Discord webhook failed: %s\nResponse: %s", resp.Status, respBody)
	}

	return nil
}

func main() {
	urls := []string{
		"https://www.bestbuy.com/site/nintendo-switch-2-mario-kart-world-bundle-nintendo-switch-2/6614325.p?skuId=6614325",
		"https://www.bestbuy.com/site/nintendo-switch-2-system-nintendo-switch-2/6614313.p?skuId=6614313",
		// "https://www.bestbuy.com/site/sony-playstation-5-dualsense-wireless-controller-midnight-black/6464307.p?skuId=6464307",
	}

	for {
		for _, url := range urls {
			available, title, err := checkInventory(url)
			if err != nil {
				log.Printf("Error checking inventory: %v\n", err)
			} else if available {
				log.Printf("%s is in stock! Sending alert...\n", title)
				message := fmt.Sprintf("%s is now in stock!\n\nBuy now: %s", title, url)

				if err := sendDiscordAlert(message); err != nil {
					log.Printf("Error sending alert to discord: %v\n", err)
				} else {
					log.Println("Notification sent.")
				}
			} else {
				log.Printf("%s is still out of stock.\n", title)
			}
		}
		time.Sleep(30 * time.Second)
	}
}
