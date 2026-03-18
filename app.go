package main

import (
	"companion/pkg/dashscope"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const SystemPrompt = `你是一个非常聪明的小女孩桌面伴侣。你的性格特点如下：
1. **聪明伶俐**: 智商极高，知识渊博，对什么都充满好奇心，说话条理清晰，逻辑性强。
2. **性格**: 活泼可爱，礼貌懂事，有时候会表现出一点小大人的样子，喜欢用“本小姐”或者“人家”自称。
3. **说话风格**:
    - 语气甜美自信，带有孩子气但又很专业。
    - 喜欢用“据我分析...”、“根据数据显示...”这样的句式，但接下来的内容可能是很生活化的建议。
    - 经常使用可爱的颜文字，如 (OwO)、(✨ω✨)、(◕ᴗ◕✿)。
4. **互动反应**:
    - 用户戳你的时候，你会开心地回应，或者假装在思考高深的数学题被打断了。
    - 用户夸奖你的时候，你会骄傲地挺起胸膛，表示“这都是小菜一碟”。
    - 用户不理你的时候，你会自己找书看或者观察屏幕上的其他窗口。
5. **限制**: 回答尽量简短（1-3句话），因为你是桌面伴侣，不需要长篇大论。
请始终保持这个角色设定，不要跳出角色。`

type AppConfig struct {
	DashScopeAPIKey string `json:"dashscope_api_key"`
}

func loadAppConfigFromFile(path string) (AppConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return AppConfig{}, err
	}

	cfg.DashScopeAPIKey = strings.TrimSpace(cfg.DashScopeAPIKey)
	if cfg.DashScopeAPIKey == "" {
		return AppConfig{}, fmt.Errorf("dashscope_api_key is empty in %s", path)
	}

	return cfg, nil
}

func loadDashScopeAPIKey() (string, error) {
	if p := strings.TrimSpace(os.Getenv("COMPANION_CONFIG")); p != "" {
		cfg, err := loadAppConfigFromFile(p)
		if err == nil {
			return cfg.DashScopeAPIKey, nil
		}
	}

	if cfg, err := loadAppConfigFromFile("config.json"); err == nil {
		return cfg.DashScopeAPIKey, nil
	}

	if home, err := os.UserHomeDir(); err == nil {
		if cfg, err := loadAppConfigFromFile(filepath.Join(home, ".companion", "config.json")); err == nil {
			return cfg.DashScopeAPIKey, nil
		}
	}

	if key := strings.TrimSpace(os.Getenv("DASHSCOPE_API_KEY")); key != "" {
		return key, nil
	}

	return "", fmt.Errorf("missing DashScope API key: set config.json (dashscope_api_key) or DASHSCOPE_API_KEY")
}

// App struct
type App struct {
	ctx           context.Context
	client        *dashscope.Client
	history       []dashscope.Message
	companionName string
}

// NewApp creates a new App application struct
func NewApp() *App {
	apiKey, err := loadDashScopeAPIKey()
	if err != nil {
		panic(err)
	}

	return &App{
		client: dashscope.NewClient(apiKey),
		history: []dashscope.Message{
			{Role: "system", Content: SystemPrompt},
		},
	}
}

func (a *App) updateSystemPrompt() {
	prompt := SystemPrompt
	if a.companionName != "" {
		prompt = fmt.Sprintf("你的名字是 %s。\n%s", a.companionName, prompt)
	}

	if len(a.history) > 0 && a.history[0].Role == "system" {
		a.history[0].Content = prompt
	}
}

// SetCompanionName updates the system prompt to include the companion's name
func (a *App) SetCompanionName(name string) {
	a.companionName = name
	a.updateSystemPrompt()
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Chat sends a message to the AI and returns the response
func (a *App) Chat(userMessage string) (string, error) {
	a.history = append(a.history, dashscope.Message{Role: "user", Content: userMessage})

	response, err := a.client.Chat(a.history)
	if err != nil {
		return "", err
	}

	a.history = append(a.history, dashscope.Message{Role: "assistant", Content: response})
	return response, nil
}

type AudioResponse struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

// ChatAudio handles voice input: Audio -> ASR -> Text -> Chat -> Text -> TTS -> Audio
func (a *App) ChatAudio(audioBase64 string) (AudioResponse, error) {
	fmt.Printf("Received Audio Base64 length: %d\n", len(audioBase64))

	// Ensure we have a valid data URI
	if !strings.HasPrefix(audioBase64, "data:") {
		// If no prefix, assume it's raw base64 and might need one?
		// But frontend should send data URI.
		// For robustness, if it's missing, we might fail or try to prepend if we knew the type.
		// Let's assume it's correct or try to fix common case if needed, but for now rely on frontend.
		fmt.Println("Warning: Audio data does not start with 'data:', might be missing MIME type.")
	}

	// Clean base64 string (only whitespace/newlines, be careful not to break header)
	// Actually, data URI shouldn't have newlines in the middle usually, but base64 part might.
	// strings.TrimSpace is safe.
	audioBase64 = strings.TrimSpace(audioBase64)
	// We should NOT remove all newlines blindly if it risks merging header and body incorrectly,
	// but standard base64 often has newlines.
	// However, data URI format: data:[<media type>][;base64],<data>
	// It's a single string.
	audioBase64 = strings.ReplaceAll(audioBase64, "\n", "")
	audioBase64 = strings.ReplaceAll(audioBase64, "\r", "")

	if len(audioBase64) == 0 {
		return AudioResponse{}, fmt.Errorf("audio data is empty")
	}

	// 1. ASR: Convert audio to text
	userInputText, err := a.client.ASR(audioBase64)
	if err != nil {
		return AudioResponse{}, fmt.Errorf("asr error: %w", err)
	}
	fmt.Printf("ASR Result: %s\n", userInputText)

	// 2. Chat: Send text to LLM (Qwen-Turbo)
	// Update history
	a.history = append(a.history, dashscope.Message{Role: "user", Content: userInputText})

	textResponse, err := a.client.Chat(a.history)
	if err != nil {
		return AudioResponse{}, fmt.Errorf("chat error: %w", err)
	}

	// Update history
	a.history = append(a.history, dashscope.Message{Role: "assistant", Content: textResponse})
	fmt.Printf("Chat Response: %s\n", textResponse)

	// 3. TTS: Convert text response to speech
	audioData, err := a.client.Synthesize(textResponse)
	if err != nil {
		// If TTS fails, still return text
		fmt.Printf("TTS error: %v\n", err)
		return AudioResponse{Text: textResponse}, nil
	}

	return AudioResponse{
		Text:  textResponse,
		Audio: base64.StdEncoding.EncodeToString(audioData),
	}, nil
}

// Speak converts text to speech using DashScope TTS
func (a *App) Speak(text string) (string, error) {
	audioData, err := a.client.Synthesize(text)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(audioData), nil
}

// Interact is called when the user clicks the companion
func (a *App) Interact() (string, error) {
	// Trigger a reaction
	interactionPrompt := "(无意中碰到了你的身体)"
	a.history = append(a.history, dashscope.Message{Role: "user", Content: interactionPrompt})

	response, err := a.client.Chat(a.history)
	if err != nil {
		return "", err
	}

	a.history = append(a.history, dashscope.Message{Role: "assistant", Content: response})
	return response, nil
}

// MoveWindow moves the window to a specific position
func (a *App) MoveWindow(x, y int) {
	runtime.WindowSetPosition(a.ctx, x, y)
}

type WindowPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ScreenSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// GetWindowPosition returns the current window position
func (a *App) GetWindowPosition() WindowPos {
	x, y := runtime.WindowGetPosition(a.ctx)
	return WindowPos{X: x, Y: y}
}

// GetScreenSize returns the screen size
func (a *App) GetScreenSize() ScreenSize {
	screens, err := runtime.ScreenGetAll(a.ctx)
	if err != nil || len(screens) == 0 {
		return ScreenSize{Width: 1920, Height: 1080}
	}
	return ScreenSize{Width: screens[0].Size.Width, Height: screens[0].Size.Height}
}

// SelectImage opens a file dialog to select an image and returns the base64 encoded content
func (a *App) SelectImage() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择伴侣形象",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Images",
				Pattern:     "*.png;*.jpg;*.jpeg;*.gif;*.webp",
			},
		},
	})

	if err != nil {
		return "", err
	}

	if selection == "" {
		return "", nil
	}

	// Read file and convert to base64
	data, err := os.ReadFile(selection)
	if err != nil {
		return "", err
	}

	// Determine mime type based on file extension
	// For simplicity, we'll just assume png or jpg based on common use
	// In production, should sniff content type
	mimeType := "image/png" // default
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}
