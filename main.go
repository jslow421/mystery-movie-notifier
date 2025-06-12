package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Config struct {
	TargetURL       string `json:"target_url"`
	ElementSelector string `json:"element_selector"`
	SMSAPI          string `json:"sms_api"`
	APIKey          string `json:"api_key"`
	PhoneNumber     string `json:"phone_number"`
	DatabaseURL     string `json:"database_url"`
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := NewDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Setup Chrome context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Scrape the website
	text, err := scrapeWebsite(ctx, config.TargetURL, config.ElementSelector)
	if err != nil {
		log.Fatal("Failed to scrape website:", err)
	}

	fmt.Printf("Scraped text: %s\n", text)

	// Extract date from the text
	currentDate, err := extractDate(text)
	if err != nil {
		log.Fatal("Failed to extract date:", err)
	}

	fmt.Printf("Extracted date: %s\n", currentDate.Format("2006-01-02"))

	// Get the latest date from database
	previousDate, err := db.GetLatestDate()
	if err != nil {
		log.Printf("Failed to get previous date: %v", err)
		previousDate = time.Time{}
	}

	if currentDate.After(previousDate) {
		fmt.Println("New date detected, sending SMS notification")
		
		err = sendSMSNotification(config, text, currentDate)
		notificationSent := err == nil
		
		if err != nil {
			log.Printf("Failed to send SMS: %v", err)
		} else {
			fmt.Println("SMS sent successfully")
		}

		// Save the notification record to database
		err = db.SaveNotification(text, currentDate, notificationSent)
		if err != nil {
			log.Printf("Failed to save notification to database: %v", err)
		}
	} else {
		fmt.Println("No new date detected")
		
		// Still save to database for tracking purposes
		err = db.SaveNotification(text, currentDate, false)
		if err != nil {
			log.Printf("Failed to save notification to database: %v", err)
		}
	}
}

func loadConfig() (*Config, error) {
	config := &Config{
		TargetURL:       getEnvOrDefault("TARGET_URL", ""),
		ElementSelector: getEnvOrDefault("ELEMENT_SELECTOR", ""),
		SMSAPI:          getEnvOrDefault("SMS_API", ""),
		APIKey:          getEnvOrDefault("API_KEY", ""),
		PhoneNumber:     getEnvOrDefault("PHONE_NUMBER", ""),
		DatabaseURL:     getEnvOrDefault("DATABASE_URL", ""),
	}

	if config.TargetURL == "" || config.ElementSelector == "" || config.DatabaseURL == "" {
		return nil, fmt.Errorf("TARGET_URL, ELEMENT_SELECTOR, and DATABASE_URL are required")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func scrapeWebsite(ctx context.Context, url, selector string) (string, error) {
	var text string
	
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(text), nil
}

func extractDate(text string) (time.Time, error) {
	// Common date formats to try
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
		"02-01-2006",
		"2006/01/02",
	}

	// Try to find and parse dates in the text
	for _, format := range formats {
		if date, err := time.Parse(format, text); err == nil {
			return date, nil
		}
	}

	// If no exact match, try to find date-like patterns
	words := strings.Fields(text)
	for i, word := range words {
		for _, format := range formats {
			// Try current word
			if date, err := time.Parse(format, word); err == nil {
				return date, nil
			}
			
			// Try combining with next words for multi-word dates
			if i < len(words)-1 {
				combined := word + " " + words[i+1]
				if date, err := time.Parse(format, combined); err == nil {
					return date, nil
				}
			}
			
			if i < len(words)-2 {
				combined := word + " " + words[i+1] + " " + words[i+2]
				if date, err := time.Parse(format, combined); err == nil {
					return date, nil
				}
			}
		}
	}

	return time.Time{}, fmt.Errorf("no date found in text: %s", text)
}


func sendSMSNotification(config *Config, text string, date time.Time) error {
	if config.SMSAPI == "" {
		return fmt.Errorf("SMS API not configured")
	}

	message := fmt.Sprintf("Mystery Movie Notifier: New date detected! %s - %s", 
		date.Format("January 2, 2006"), text)

	// Determine SMS provider and send accordingly
	switch {
	case strings.Contains(config.SMSAPI, "twilio"):
		accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
		authToken := os.Getenv("TWILIO_AUTH_TOKEN")
		fromNumber := os.Getenv("TWILIO_FROM_NUMBER")
		
		if accountSID == "" || authToken == "" || fromNumber == "" {
			return fmt.Errorf("Twilio credentials not configured")
		}
		
		return sendTwilioSMS(accountSID, authToken, fromNumber, config.PhoneNumber, message)
		
	case strings.Contains(config.SMSAPI, "webhook"):
		return sendWebhookSMS(config.SMSAPI, message, config.PhoneNumber)
		
	default:
		// Fallback to webhook
		return sendWebhookSMS(config.SMSAPI, message, config.PhoneNumber)
	}
}