package dashscope

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	BaseURL  = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"
	TTSURL   = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	ASRURL   = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	Model    = "qwen-turbo"
	TTSModel = "qwen3-tts-flash"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Input struct {
	Messages []Message `json:"messages"`
}

type Parameters struct {
	ResultFormat string `json:"result_format"`
}

type Request struct {
	Model      string     `json:"model"`
	Input      Input      `json:"input"`
	Parameters Parameters `json:"parameters"`
}

type Output struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

type Response struct {
	Output Output `json:"output"`
	Usage  struct {
		OutputTokens int `json:"output_tokens"`
		InputTokens  int `json:"input_tokens"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

// OpenAI Compatible ASR Structures
type ASRContentPart struct {
	Type       string      `json:"type"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
}

type InputAudio struct {
	Data string `json:"data"`
}

type ASRCompatibleMessage struct {
	Role    string           `json:"role"`
	Content []ASRContentPart `json:"content"`
}

type ASRCompatibleRequest struct {
	Model    string                 `json:"model"`
	Messages []ASRCompatibleMessage `json:"messages"`
}

type ASRCompatibleResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Code    interface{} `json:"code"`
	} `json:"error,omitempty"`
}

type Client struct {
	APIKey string
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
	}
}

func (c *Client) Chat(messages []Message) (string, error) {
	reqBody := Request{
		Model: Model,
		Input: Input{
			Messages: messages,
		},
		Parameters: Parameters{
			ResultFormat: "text",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Debug log: Print response
	fmt.Printf("Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Code != "" {
		return "", fmt.Errorf("api error: %s - %s", response.Code, response.Message)
	}

	return response.Output.Text, nil
}

func (c *Client) ASR(audioBase64 string) (string, error) {
	reqBody := ASRCompatibleRequest{
		Model: "qwen3-asr-flash",
		Messages: []ASRCompatibleMessage{
			{
				Role: "user",
				Content: []ASRContentPart{
					{
						Type: "input_audio",
						InputAudio: &InputAudio{
							Data: audioBase64,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug log
	debugJSON := string(jsonData)
	if len(debugJSON) > 1000 {
		fmt.Printf("ASR Request Payload: %s... (truncated)\n", debugJSON[:200])
	} else {
		fmt.Printf("ASR Request Payload: %s\n", debugJSON)
	}
	fmt.Printf("ASR Request URL: %s\n", ASRURL)

	req, err := http.NewRequest("POST", ASRURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug log
	fmt.Printf("ASR Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("asr request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ASRCompatibleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("asr api error: %s - %s", response.Error.Code, response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no recognition result")
	}

	return response.Choices[0].Message.Content, nil
}

type TTSInput struct {
	Text string `json:"text"`
}

type TTSParameters struct {
	TextType string `json:"text_type,omitempty"`
	Format   string `json:"format,omitempty"`
	Voice    string `json:"voice,omitempty"`
}

type TTSRequest struct {
	Model      string        `json:"model"`
	Input      TTSInput      `json:"input"`
	Parameters TTSParameters `json:"parameters,omitempty"`
}

type TTSResponse struct {
	Output struct {
		Audio struct {
			URL string `json:"url"`
		} `json:"audio"`
		FinishReason string `json:"finish_reason"`
	} `json:"output"`
	Usage struct {
		Characters int `json:"characters"`
	} `json:"usage"`
	RequestID string `json:"request_id"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (c *Client) Synthesize(text string) ([]byte, error) {
	reqBody := TTSRequest{
		Model: TTSModel,
		Input: TTSInput{
			Text: text,
		},
		Parameters: TTSParameters{
			// TextType: "PlainText", // qwen-tts-flash usually infers this or uses specific params
			// Format:   "mp3",       // qwen-tts-flash output format is usually specified differently or defaults
			// According to docs for qwen3-tts-flash, basic usage doesn't strictly need these if defaults work,
			// but let's check if we need to set voice.
			// The example uses "Cherry".
			Voice: "Cherry",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug log for TTS Request
	fmt.Printf("[TTS Step] Request URL: %s\n", TTSURL)
	debugJSON := string(jsonData)
	if len(debugJSON) > 1000 {
		fmt.Printf("[TTS Step] Request Payload: %s... (truncated)\n", debugJSON[:200])
	} else {
		fmt.Printf("[TTS Step] Request Payload: %s\n", debugJSON)
	}

	req, err := http.NewRequest("POST", TTSURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("X-DashScope-Data-Inspection", "enable") // Optional, but good practice

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[TTS Step] Error Response Body: %s\n", string(body))
		return nil, fmt.Errorf("tts request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Debug log
	fmt.Printf("[TTS Step] Response Body: %s\n", string(body))

	var response TTSResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Code != "" {
		return nil, fmt.Errorf("tts api error: %s - %s", response.Code, response.Message)
	}

	if response.Output.Audio.URL == "" {
		return nil, fmt.Errorf("no audio url in response")
	}

	audioURL := response.Output.Audio.URL
	fmt.Printf("[TTS Step] Got Audio URL: %s\n", audioURL)

	// Download the audio
	audioResp, err := http.Get(audioURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer audioResp.Body.Close()

	if audioResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download audio, status: %d", audioResp.StatusCode)
	}

	audioData, err := io.ReadAll(audioResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded audio data: %w", err)
	}

	fmt.Printf("[TTS Step] Success! Downloaded audio data size: %d bytes\n", len(audioData))
	return audioData, nil
}
