package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	a := app.New()
	w := a.NewWindow("Text Editor with ChatGPT")
	w.Resize(fyne.NewSize(800, 600))

	content := widget.NewMultiLineEntry()
	content.Wrapping = fyne.TextWrapWord

	content.OnTypedKey = func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyF1 {
			go func() {
				resp, err := getChatGPTResponse(content.Text, apiKey)
				if err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Error",
						Content: "Failed to fetch response from ChatGPT: " + err.Error(),
					})
					return
				}
				fyne.CurrentApp().Queue(func() {
					content.SetText(content.Text + resp)
				})
			}()
		}
	}

	w.SetContent(container.NewScroll(content))
	w.ShowAndRun()
}

func getChatGPTResponse(prompt, apiKey string) (string, error) {
	payload := map[string]interface{}{
		"prompt":     prompt,
		"max_tokens": 50,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/text-davinci-003/completions", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var responseObj map[string][]map[string]interface{}
	json.Unmarshal(responseBytes, &responseObj)
	if len(responseObj["choices"]) > 0 && len(responseObj["choices"][0]["text"].(string)) > 0 {
		return responseObj["choices"][0]["text"].(string), nil
	}
	return "", nil
}
