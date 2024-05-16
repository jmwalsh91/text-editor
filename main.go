package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
)

var threadID string // global variable to hold the thread ID

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	myApp := app.New()
	myWindow := myApp.NewWindow("Text Editor with ChatGPT")
	myWindow.Resize(fyne.NewSize(800, 400))

	// Initialize the thread at application start
	var initErr error
	threadID, initErr = initializeThread(apiKey)
	if initErr != nil {
		log.Println("Failed to initialize thread:", initErr)
	} else {
		log.Println("Thread initialized with ID:", threadID)
	}

	content := widget.NewMultiLineEntry()
	content.SetPlaceHolder("Enter text here...")

	chatButton := widget.NewButton("Send to ChatGPT", func() {
		if threadID == "" {
			dialog.ShowError(fmt.Errorf("thread not initialized"), myWindow)
			return
		}
		response, err := sendMessageAndGetResponse(apiKey, threadID, content.Text)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		content.SetText(content.Text + " " + response)
	})

	myWindow.SetContent(container.NewVBox(content, chatButton))
	myWindow.ShowAndRun()
}

func initializeThread(apiKey string) (string, error) {
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/threads", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error making API request to initialize thread:", err)
		return "as", err
	}

	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body during thread initialization:", err)
		return "", err
	}

	var responseObj map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseObj); err != nil {
		log.Println("Error unmarshaling response during thread initialization:", err)
		return "", err
	}

	if id, ok := responseObj["id"].(string); ok {
		return id, nil
	} else {
		return "", fmt.Errorf("failed to retrieve thread ID: %v", responseObj)
	}
}

func sendMessageAndGetResponse(apiKey, threadID, message string) (string, error) {
	payload := map[string]interface{}{
		"thread_id": threadID,
		"messages": []map[string]string{
			{"role": "user", "content": message},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling message request:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", threadID), bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error making API request to send message:", err)
		return "", err
	}
	defer resp.Body.Close()

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body from message request:", err)
		return "", err
	}

	var responseObj map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseObj); err != nil {
		log.Println("Error unmarshaling response from message request:", err)
		return "", err
	}

	// Extracting response from OpenAI
	messages := responseObj["messages"].([]interface{})
	lastResponse := messages[len(messages)-1].(map[string]interface{})
	text, ok := lastResponse["content"].(string)
	if !ok {
		return "", fmt.Errorf("failed to extract response text")
	}
	return text, nil
}
