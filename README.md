# Desktop Companion（桌面 AI 伴侣）

[English README](README.en.md)

基于 Wails（Go）+ Vue + Vite 的桌面伴侣应用，集成阿里云百炼（DashScope / Qwen）能力，支持文字对话与语音对话（ASR + TTS）。

## 功能
- 桌面伴侣：透明窗体、始终置顶、可拖拽移动
- 文字对话：调用 Qwen 模型进行回复
- 语音对话：录音转文字（ASR）→ 模型回复 → 文字转语音（TTS）并播放
- 形象与人设：可切换/替换头像；系统提示词可自定义

## 技术栈
- 后端：Go 1.23 + Wails v2
- 前端：Vue 3 + Vite
- 模型服务：DashScope（Qwen）
  - 聊天：`qwen-turbo`
  - ASR：`qwen3-asr-flash`（OpenAI 兼容接口）
  - TTS：`qwen3-tts-flash`（多模态生成接口，返回音频 URL）

## 目录结构
- [app.go](app.go)：应用业务逻辑（Chat / ChatAudio / Speak / SelectImage 等）
- [pkg/dashscope/client.go](pkg/dashscope/client.go)：DashScope API 客户端（Chat / ASR / TTS）
- [frontend](frontend/)：Vue 前端
- [config.json](config.json)：本地配置（已被 .gitignore 忽略，请勿提交）
- [config.example.json](config.example.json)：配置模板（可提交）

## 快速开始

### 1) 安装依赖
- 安装 Go 与 Node.js
- 安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 2) 配置密钥（必做）

复制模板并填入你的 DashScope API Key：

```bash
cp config.example.json config.json
```

`config.json` 格式：

```json
{
  "dashscope_api_key": "sk-xxxx"
}
```

密钥读取优先级（从高到低）：
- `COMPANION_CONFIG` 指定的配置文件路径
- 项目根目录 `./config.json`
- `~/.companion/config.json`
- 环境变量 `DASHSCOPE_API_KEY`

### 3) 开发运行

在项目根目录执行：

```bash
wails dev
```

### 4) 打包构建

```bash
wails build
```

## 个性化
- 更换头像：替换 [frontend/src/assets/images](frontend/src/assets/images/) 下资源，或在界面中选择
- 修改人设：调整 [app.go](app.go) 中 `SystemPrompt`

## 常见问题

### 语音输入（ASR）无结果/报错
- ASR 使用 OpenAI 兼容接口，`input_audio.data` 需要是 Data URI 格式：`data:<mime>;base64,<base64>`
- 前端录音一般为 `audio/webm` 或 `audio/mp4`（不同浏览器/系统会不同）

### 语音输出（TTS）播放失败
- TTS 接口会返回音频 URL，后端会自动下载音频数据并返回给前端播放
- 如果下载失败，优先检查网络与 URL 是否过期（`expires_at`）

## 安全与 Git
- `config.json` / `.env*` 已在 [.gitignore](.gitignore) 中忽略，避免提交密钥
- `build/` 已整体忽略
