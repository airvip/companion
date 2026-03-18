# Desktop Companion

[中文 README](README.md)

A desktop AI companion built with Wails (Go) + Vue + Vite, integrated with Alibaba Cloud Bailian (DashScope / Qwen). It supports both text chat and voice chat (ASR + TTS).

## Features
- Desktop companion: transparent window, always-on-top, draggable
- Text chat: uses Qwen model to generate responses
- Voice chat: Audio → ASR → Text → Chat → Text → TTS → Audio playback
- Customizable: switch/replace avatars; customize system prompt/personality

## Tech Stack
- Backend: Go 1.23 + Wails v2
- Frontend: Vue 3 + Vite
- Model provider: DashScope (Qwen)
  - Chat: `qwen-turbo`
  - ASR: `qwen3-asr-flash` (OpenAI-compatible endpoint)
  - TTS: `qwen3-tts-flash` (multimodal generation endpoint; returns audio URL)

## Project Structure
- [app.go](app.go): app logic (Chat / ChatAudio / Speak / SelectImage, etc.)
- [pkg/dashscope/client.go](pkg/dashscope/client.go): DashScope client (Chat / ASR / TTS)
- [frontend](frontend/): Vue UI
- [config.json](config.json): local config (gitignored; do not commit)
- [config.example.json](config.example.json): config template (safe to commit)

## Getting Started

### 1) Prerequisites
- Install Go and Node.js
- Install Wails CLI:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 2) Configure API Key (required)

Copy the template and fill in your DashScope API key:

```bash
cp config.example.json config.json
```

`config.json` format:

```json
{
  "dashscope_api_key": "sk-xxxx"
}
```

API key lookup order (highest priority first):
- `COMPANION_CONFIG` points to a config file path
- Project root `./config.json`
- `~/.companion/config.json`
- Environment variable `DASHSCOPE_API_KEY`

### 3) Run in development

From the project root:

```bash
wails dev
```

### 4) Build

```bash
wails build
```

## Customization
- Avatar: replace assets under [frontend/src/assets/images](frontend/src/assets/images/) or select in UI
- Personality: edit `SystemPrompt` in [app.go](app.go)

## Troubleshooting

### ASR returns empty / fails
- The ASR request uses OpenAI-compatible format. `input_audio.data` must be a Data URI: `data:<mime>;base64,<base64>`
- The frontend recorder may produce `audio/webm` or `audio/mp4` depending on OS/browser runtime

### TTS playback fails
- TTS returns an audio URL; backend downloads the audio bytes and returns them for playback
- If download fails, check network access and whether the URL is expired (`expires_at`)

## Security & Git
- `config.json` and `.env*` are ignored in [.gitignore](.gitignore) to prevent leaking secrets
- `build/` is fully ignored

