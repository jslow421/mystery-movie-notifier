package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// TwilioSMS sends SMS via Twilio API
func sendTwilioSMS(accountSID, authToken, fromNumber, toNumber, message string) error {
	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)
	
	msgData := url.Values{}
	msgData.Set("To", toNumber)
	msgData.Set("From", fromNumber)
	msgData.Set("Body", message)
	msgDataReader := strings.NewReader(msgData.Encode())
	
	req, err := http.NewRequest("POST", urlStr, msgDataReader)
	if err != nil {
		return err
	}
	req.SetBasicAuth(accountSID, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	
	return fmt.Errorf("SMS API returned status %d", resp.StatusCode)
}

// GenericWebhookSMS sends SMS via a generic webhook
func sendWebhookSMS(webhookURL, message, phoneNumber string) error {
	payload := map[string]interface{}{
		"message": message,
		"to":      phoneNumber,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	
	return fmt.Errorf("Webhook returned status %d", resp.StatusCode)
}