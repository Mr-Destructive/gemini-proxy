package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GeminiAPI struct {
	BaseURL        string
	Endpoint       string
	Timeout        time.Duration
	Retries        int
	ConversationID string
	ResponseID     string
	HTTPClient     *http.Client
}

// NewGeminiAPI creates a new GeminiAPI instance with default settings
func NewGeminiAPI() *GeminiAPI {
	return &GeminiAPI{
		BaseURL:  "https://gemini.google.com",
		Endpoint: "/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate",
		Timeout:  60 * time.Second,
		Retries:  3,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewGeminiAPIWithConfig creates a GeminiAPI with custom timeout and retries
func NewGeminiAPIWithConfig(timeout time.Duration, retries int) *GeminiAPI {
	api := NewGeminiAPI()
	api.Timeout = timeout
	api.Retries = retries
	api.HTTPClient.Timeout = timeout
	return api
}

// buildPayload constructs the request payload
func (g *GeminiAPI) buildPayload(message string) map[string]string {
	inner := []interface{}{
		[]interface{}{message, 0, nil, nil, nil, nil, 0},
		[]string{"en-US"},
		[]interface{}{g.ConversationID, g.ResponseID},
	}

	innerJSON, _ := json.Marshal(inner)
	outer := []interface{}{nil, string(innerJSON)}
	outerJSON, _ := json.Marshal(outer)

	return map[string]string{
		"f.req": string(outerJSON),
	}
}

// parseResponse extracts the response text from API response
func (g *GeminiAPI) parseResponse(text string) string {
	if strings.HasPrefix(text, ")]}'") {
		text = text[4:]
	}

	var data []interface{}
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) == 0 {
		return ""
	}

	if err := json.Unmarshal([]byte(lines[0]), &data); err != nil {
		return ""
	}

	for _, item := range data {
		itemArr, ok := item.([]interface{})
		if !ok || len(itemArr) < 3 {
			continue
		}

		if itemArr[0] != "wrb.fr" || itemArr[2] == nil {
			continue
		}

		innerStr, ok := itemArr[2].(string)
		if !ok {
			continue
		}

		var inner []interface{}
		if err := json.Unmarshal([]byte(innerStr), &inner); err != nil {
			continue
		}

		if len(inner) <= 4 {
			continue
		}

		content, ok := inner[4].([]interface{})
		if !ok || len(content) == 0 {
			continue
		}

		for _, part := range content {
			partArr, ok := part.([]interface{})
			if !ok || len(partArr) < 2 {
				continue
			}

			var contentText string
			switch v := partArr[1].(type) {
			case string:
				contentText = v
			case []interface{}:
				if len(v) > 0 {
					if s, ok := v[0].(string); ok {
						contentText = s
					}
				}
			}

			if contentText != "" {
				if strings.Contains(contentText, "```") {
					segments := strings.Split(contentText, "```")
					if len(segments) >= 3 {
						return strings.TrimSpace(segments[len(segments)-1])
					}
				} else {
					return contentText
				}
			}
		}
	}

	return ""
}

// Ask sends a message to Gemini and returns the response
func (g *GeminiAPI) Ask(message string) (string, error) {
	apiURL := g.BaseURL + g.Endpoint
	payload := g.buildPayload(message)

	// Encode payload as form-urlencoded
	formData := url.Values{}
	for k, v := range payload {
		formData.Set(k, v)
	}
	bodyStr := formData.Encode()

	for attempt := 0; attempt < g.Retries; attempt++ {
		req, _ := http.NewRequest("POST", apiURL, bytes.NewBufferString(bodyStr))
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
		req.Header.Set("Origin", g.BaseURL)
		req.Header.Set("Referer", g.BaseURL+"/")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

		resp, err := g.HTTPClient.Do(req)
		if err != nil {
			if attempt < g.Retries-1 {
				wait := time.Duration((attempt+1)*2) * time.Second
				fmt.Printf("Timeout, retrying in %ds... (attempt %d/%d)\n", wait/time.Second, attempt+1, g.Retries)
				time.Sleep(wait)
				continue
			}
			return "", fmt.Errorf("API error after %d retries: %w", g.Retries, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("status %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		result := g.parseResponse(string(body))
		if result != "" {
			return result, nil
		}

		if attempt < g.Retries-1 {
			time.Sleep(time.Second)
		}
	}

	return "", fmt.Errorf("empty response")
}

// ClearConversation resets conversation state
func (g *GeminiAPI) ClearConversation() {
	g.ConversationID = ""
	g.ResponseID = ""
}
