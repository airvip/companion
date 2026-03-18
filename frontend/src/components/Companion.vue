<script setup>
import {reactive, ref, onMounted} from 'vue'
import {Chat, Interact, MoveWindow, GetWindowPosition, GetScreenSize, SetCompanionName, Speak, ChatAudio} from '../../wailsjs/go/main/App'
import defaultAvatar from '../assets/images/green_dragon.gif'
import robotAvatar from '../assets/images/robot.svg'
import catAvatar from '../assets/images/cat.svg'
import ghostAvatar from '../assets/images/ghost.svg'
import girl1Avatar from '../assets/images/girl1.svg'
import girl2Avatar from '../assets/images/girl2.svg'
import acgGirl1 from '../assets/images/acg_girl1.gif'

const data = reactive({
  response: "你好呀！我是你的智能小助手。有什么难题尽管问我吧！(✨ω✨)",
  input: "",
  avatarUrl: acgGirl1
})

const isTalking = ref(false)
const isVoiceMode = ref(false)
const isRecording = ref(false)
let mediaRecorder = null
let audioChunks = []
const showInput = ref(false)
const showResponse = ref(false)
let responseTimer = null

const fileInput = ref(null)
const showMenu = ref(false)
const showAvatarSelector = ref(false)
const showTimer = ref(false)
const showSettings = ref(false)
const showNameSetting = ref(false)
const showCompanionNameSetting = ref(false)
const companionName = ref("")
const timerDuration = ref(15) // minutes
const timerMessage = ref("") // reminder message
const remainingTime = ref(0)
const timerInterval = ref(null)

// Interaction Settings
const settings = reactive({
    shakeInterval: 30, // minutes
    feedInterval: 4, // hours
    moveInterval: 60, // minutes
    lastInteractTime: Date.now(),
    lastFeedTime: Date.now(),
})

const builtInAvatars = [
    { name: '默认', url: defaultAvatar },
    { name: '机器人', url: robotAvatar },
    { name: '猫咪', url: catAvatar },
    { name: '幽灵', url: ghostAvatar },
    { name: '金发少女', url: girl1Avatar },
    { name: '猫耳少女', url: girl2Avatar },
    { name: 'ACG少女', url: acgGirl1 },
]

// Background checks
let checkInterval = null

onMounted(() => {
    const savedAvatar = localStorage.getItem('companion-avatar')
    if (savedAvatar) {
        data.avatarUrl = savedAvatar
    }
    
    // Load settings
    const savedSettings = localStorage.getItem('companion-settings')
    if (savedSettings) {
        Object.assign(settings, JSON.parse(savedSettings))
    }

    // Load companion name
    const savedCompanionName = localStorage.getItem('companion-name')
    if (savedCompanionName) {
        companionName.value = savedCompanionName
        try {
            SetCompanionName(savedCompanionName)
        } catch (e) {
            console.error("SetCompanionName not ready yet", e)
        }
    }

    // Start background checks
    checkInterval = setInterval(checkStatus, 60000) // Check every minute

    // Setup Speech Recognition - Removed in favor of MediaRecorder + Backend ASR
    // checkInterval = setInterval(checkStatus, 60000)
})

function checkStatus() {
    const now = Date.now()
    
    // 1. Shake if no interaction
    if ((now - settings.lastInteractTime) > settings.shakeInterval * 60 * 1000) {
        shakeWindow()
        // Reset to prevent continuous shaking, or let user interaction reset it
        // Here we just shake once and maybe set a flag? 
        // For now, let's just shake. To avoid constant shaking, we could update lastInteractTime slightly?
        // But user wants "shake if > X minutes", so it should probably happen periodically until interacted.
        // Let's add a small buffer so it doesn't shake every minute if interval is small?
        // Or better: just shake.
    }

    // 2. Feeding reminder
    if ((now - settings.lastFeedTime) > settings.feedInterval * 60 * 60 * 1000) {
        data.response = "我饿了！快给我投喂零食！(双击 -> 投喂)"
        showResponse.value = true
        showInput.value = false // Just show bubble
        scheduleAutoHide()
        isTalking.value = true
    }

    // 3. Move slowly if no interaction
    if ((now - settings.lastInteractTime) > settings.moveInterval * 60 * 1000) {
        moveRandomly()
    }
}

async function shakeWindow() {
    try {
        const pos = await GetWindowPosition()
        const startX = pos.x
        const startY = pos.y
        const intensity = 10
        const duration = 500
        const step = 50
        
        const startTime = Date.now()
        
        const shake = () => {
            const elapsed = Date.now() - startTime
            if (elapsed > duration) {
                MoveWindow(startX, startY)
                return
            }
            
            const offsetX = (Math.random() - 0.5) * intensity
            const offsetY = (Math.random() - 0.5) * intensity
            MoveWindow(Math.floor(startX + offsetX), Math.floor(startY + offsetY))
            
            setTimeout(shake, step)
        }
        
        shake()
    } catch (e) {
        console.error("Shake error:", e)
    }
}

    async function moveRandomly() {
    try {
        const pos = await GetWindowPosition()
        const screen = await GetScreenSize()
        
        // Move by larger amount for full screen effect
        const step = 50 
        // Random direction and distance
        const dx = (Math.random() - 0.5) * step * 2 // -50 to 50
        const dy = (Math.random() - 0.5) * step * 2 // -50 to 50
        
        // Occasionally make a big jump to "walk" across screen
        const bigJump = Math.random() > 0.8
        const jumpX = bigJump ? (Math.random() - 0.5) * 200 : 0
        const jumpY = bigJump ? (Math.random() - 0.5) * 200 : 0

        let newX = pos.x + dx + jumpX
        let newY = pos.y + dy + jumpY
        
        // Boundary check (keep fully within screen)
        // Assume window size is approx 300x400 max
        const winWidth = 300
        const winHeight = 400
        
        if (newX < 0) newX = 0
        if (newY < 0) newY = 0
        if (newX > screen.width - winWidth) newX = screen.width - winWidth
        if (newY > screen.height - winHeight) newY = screen.height - winHeight
        
        MoveWindow(Math.floor(newX), Math.floor(newY))
    } catch (e) {
        console.error("Move error:", e)
    }
}

function openSettings() {
    showMenu.value = false
    showSettings.value = true
}

function openCompanionNameSetting() {
    showMenu.value = false
    showCompanionNameSetting.value = true
}

function saveCompanionName() {
    if (companionName.value.trim()) {
        localStorage.setItem('companion-name', companionName.value)
        try {
            SetCompanionName(companionName.value)
        } catch (e) {
            console.error(e)
        }
        data.response = `嗯嗯！我现在的名字是 ${companionName.value}，请多指教哦！`
        showCompanionNameSetting.value = false
        showResponse.value = true
        scheduleAutoHide()
        isTalking.value = true
        setTimeout(() => isTalking.value = false, 3000)
    }
}

function saveSettings() {
    localStorage.setItem('companion-settings', JSON.stringify(settings))
    showSettings.value = false
}

function feedCompanion() {
    settings.lastFeedTime = Date.now()
    data.response = "啊呜！好吃！(幸福地眯起了眼睛)"
    isTalking.value = true
    showMenu.value = false // Close menu
    showResponse.value = true // Show bubble
    showInput.value = false // Hide input
    saveSettings()
    scheduleAutoHide()
    
    // Stop talking animation after 3 seconds
    setTimeout(() => {
        isTalking.value = false
    }, 3000)
}

function updateInteraction() {
    settings.lastInteractTime = Date.now()
    saveSettings()
}

function changeAvatar(e) {
    if (e) e.preventDefault() // Prevent context menu
    showMenu.value = false
    showAvatarSelector.value = true
}

function handleRightClick(e) {
    e.preventDefault()
    showMenu.value = true
    showInput.value = false
    // We can keep showResponse as is, or hide it? Let's keep it if it was showing.
}

function selectAvatar(url) {
    data.avatarUrl = url
    localStorage.setItem('companion-avatar', url)
    showAvatarSelector.value = false
}

function triggerUpload() {
    fileInput.value.click()
}

function handleFileUpload(e) {
    const file = e.target.files[0]
    if (file) {
        const reader = new FileReader()
        reader.onload = (e) => {
            const result = e.target.result
            data.avatarUrl = result
            localStorage.setItem('companion-avatar', result)
            showAvatarSelector.value = false
        }
        reader.readAsDataURL(file)
    }
}

function openTimer() {
    showMenu.value = false
    showTimer.value = true
}

function startTimer() {
    remainingTime.value = timerDuration.value * 60
    showTimer.value = false
    
    if (timerInterval.value) clearInterval(timerInterval.value)
    
    timerInterval.value = setInterval(() => {
        remainingTime.value--
        if (remainingTime.value <= 0) {
            clearInterval(timerInterval.value)
            timerInterval.value = null
            const message = timerMessage.value ? `⏰ 时间到啦！${timerMessage.value}` : "⏰ 时间到啦！该休息一下了！"
            data.response = message
            isTalking.value = true
            showResponse.value = true
            // showInput.value = true // Optional: do we need input here?
            scheduleAutoHide()
            // Play a sound or maximize window if possible
        }
    }, 1000)
}

function cancelTimer() {
    if (timerInterval.value) {
        clearInterval(timerInterval.value)
        timerInterval.value = null
    }
    remainingTime.value = 0
    showTimer.value = false
}

function formatTime(seconds) {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m}:${s.toString().padStart(2, '0')}`
}

function toggleInput(e) {
  // If clicking on menu or other overlays, do not toggle input
  if (e.target.closest('.menu-overlay') || 
      e.target.closest('.avatar-selector') || 
      e.target.closest('.settings-panel') ||
      e.target.closest('.timer-setup')) {
      return
  }

  if (showMenu.value || showAvatarSelector.value || showTimer.value || showSettings.value || showCompanionNameSetting.value) {
      showMenu.value = false
      showAvatarSelector.value = false
      showTimer.value = false
      showSettings.value = false
      showCompanionNameSetting.value = false
      return
  }
  showInput.value = !showInput.value
  showResponse.value = showInput.value // Sync bubble with input toggle
  
  // Reset voice mode when closing
  if (!showInput.value) {
      isVoiceMode.value = false
      if (isRecording.value && recognition) recognition.stop()
  }

  if (showInput.value && !data.response) {
      // Optional: Trigger a greeting if opening and no previous response
      interact()
  } else if (showInput.value && data.response) {
      scheduleAutoHide() // Restart timer if re-opened
  }
}

function toggleVoiceMode() {
    isVoiceMode.value = !isVoiceMode.value
    if (!isVoiceMode.value && isRecording.value) {
        stopVoiceInput()
    }
}

async function startVoiceInput() {
    if (isRecording.value) return

    try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        
        // Prefer audio/mp4 (AAC) for better compatibility with DashScope, then webm
        let mimeType = 'audio/webm';
        if (MediaRecorder.isTypeSupported('audio/mp4')) {
            mimeType = 'audio/mp4';
        } else if (MediaRecorder.isTypeSupported('audio/webm;codecs=opus')) {
            mimeType = 'audio/webm;codecs=opus';
        }

        try {
            mediaRecorder = new MediaRecorder(stream, { mimeType })
        } catch (e) {
            console.warn(`Failed to create MediaRecorder with ${mimeType}, falling back to default`, e);
            mediaRecorder = new MediaRecorder(stream);
            // Try to guess or check actual mimeType if possible, but MediaRecorder doesn't always expose it reliably across browsers
            if (mediaRecorder.mimeType) {
                mimeType = mediaRecorder.mimeType;
            }
        }

        audioChunks = []

        mediaRecorder.ondataavailable = (event) => {
            console.log("MediaRecorder data available, size:", event.data.size)
            if (event.data.size > 0) {
                audioChunks.push(event.data)
            }
        }

        mediaRecorder.onstop = async () => {
            console.log("MediaRecorder stopped. Chunks count:", audioChunks.length)
            
            // Validate audio data
            if (audioChunks.length === 0) {
                console.warn("No audio chunks recorded.")
                data.response = "录音失败：未检测到语音数据"
                showResponse.value = true
                scheduleAutoHide()
                stream.getTracks().forEach(track => track.stop())
                return
            }

            const audioBlob = new Blob(audioChunks, { type: mimeType })
            console.log("Audio Blob size:", audioBlob.size, "type:", mimeType)
            
            if (audioBlob.size < 100) { // Too small to be valid audio
                console.warn("Audio too short/empty.")
                data.response = "录音时间太短啦"
                showResponse.value = true
                scheduleAutoHide()
                stream.getTracks().forEach(track => track.stop())
                return
            }

            const reader = new FileReader()
            reader.readAsDataURL(audioBlob)
            reader.onloadend = () => {
                const base64Audio = reader.result
                console.log("Base64 Audio prefix:", base64Audio.substring(0, 50))
                sendAudioMessage(base64Audio)
            }
            
            // Stop all tracks to release microphone
            stream.getTracks().forEach(track => track.stop())
        }

        mediaRecorder.start()
        console.log("MediaRecorder started with mimeType:", mimeType)
        isRecording.value = true
    } catch (err) {
        console.error("Error accessing microphone:", err)
        data.response = "无法访问麦克风，请检查权限"
        showResponse.value = true
        scheduleAutoHide()
    }
}

function stopVoiceInput() {
    if (isRecording.value) {
        if (mediaRecorder && mediaRecorder.state !== 'inactive') {
            mediaRecorder.stop()
        }
        isRecording.value = false
    }
}

function sendAudioMessage(base64Audio) {
    data.response = "正在思考..."
    showResponse.value = true
    isTalking.value = true

    ChatAudio(base64Audio).then(result => {
        // result is {text: "", audio: ""}
        data.response = result.text
        showResponse.value = true
        scheduleAutoHide()
        
        if (result.audio) {
            playAudio(result.audio)
        } else {
            isTalking.value = false
        }
    }).catch(err => {
        console.error("ChatAudio error:", err)
        data.response = "出错了: " + err
        showResponse.value = true
        scheduleAutoHide()
        isTalking.value = false
    })
}

function playAudio(base64) {
    try {
        const audio = new Audio("data:audio/mp3;base64," + base64)
        audio.volume = 1.0
        audio.play().catch(e => console.error("Audio play error:", e))
        isTalking.value = true
        audio.onended = () => {
            isTalking.value = false
        }
    } catch (e) {
        console.error("Audio init error:", e)
    }
}

function interact() {
  if (isTalking.value) return
  isTalking.value = true
  // Optimistic UI update or wait for response
  Interact().then(result => {
    data.response = result
    showResponse.value = true
    scheduleAutoHide()
    isTalking.value = false
  }).catch(err => {
      data.response = "Error: " + err
      showResponse.value = true
      scheduleAutoHide()
      isTalking.value = false
  })
}

function sendMessage() {
  if (!data.input) return
  
  // Allow sending even if talking, or queue it? For now just ignore isTalking check to fix "can't send" issue
  // if (isTalking.value) return 
  
  const msg = data.input
  data.input = ""
  data.response = "..."
  showResponse.value = true
  isTalking.value = true
  
  Chat(msg).then(result => {
    data.response = result
    showResponse.value = true
    scheduleAutoHide()
    
    if (isVoiceMode.value) {
        Speak(result).then(base64 => {
            if (base64) playAudio(base64)
            else isTalking.value = false
        }).catch(err => {
            console.error("Speak error:", err)
            isTalking.value = false
        })
    } else {
        isTalking.value = false
    }
  }).catch(err => {
    data.response = "Error: " + err
    showResponse.value = true
    scheduleAutoHide()
    isTalking.value = false
  })
}

function scheduleAutoHide() {
    if (responseTimer) {
        clearTimeout(responseTimer)
        responseTimer = null
    }
    
    // Calculate duration based on length
    // Base 3 seconds + 200ms per character
    const duration = Math.max(3000, data.response.length * 200)
    
    responseTimer = setTimeout(() => {
        showResponse.value = false
        data.response = "" // Clear response content
        responseTimer = null
    }, duration)
}
</script>

<template>
  <div class="companion-container">
    <div class="chat-bubble" v-if="data.response && showResponse">
      {{ data.response }}
    </div>
    
    <div class="avatar" @click="toggleInput" @contextmenu="handleRightClick" style="--wails-draggable:drag">
      <img :src="data.avatarUrl" alt="Companion" :class="{ 'talking': isTalking }"/>
    </div>

    <!-- Double Click Menu -->
    <div class="menu-overlay" v-if="showMenu" style="--wails-draggable:no-drag">
        <div class="menu-title">功能列表</div>
        <div class="menu-item" @click.stop="feedCompanion">投喂零食</div>
        <div class="menu-item" @click.stop="changeAvatar">设置形象</div>
        <div class="menu-item" @click.stop="openCompanionNameSetting">设置昵称</div>
        <div class="menu-item" @click.stop="openTimer">设置提醒</div>
        <div class="menu-item" @click.stop="openSettings">互动设置</div>
        <div class="menu-item" @click.stop="showMenu = false">关闭菜单</div>
    </div>

    <!-- Companion Name Setting -->
    <div class="settings-panel" v-if="showCompanionNameSetting">
        <div class="selector-title">设置昵称</div>
        <div class="setting-item">
            <input type="text" v-model="companionName" placeholder="请输入伴侣名字" @keyup.enter="saveCompanionName" />
        </div>
        <div class="timer-actions">
            <button @click="saveCompanionName">确定</button>
            <button @click="showCompanionNameSetting = false" class="cancel">取消</button>
        </div>
    </div>

    <!-- Interaction Settings -->
    <div class="settings-panel" v-if="showSettings">
        <div class="selector-title">互动设置</div>
        <div class="setting-item">
            <label>未互动窗口抖动 (分钟)</label>
            <input type="number" v-model="settings.shakeInterval" min="1" />
        </div>
        <div class="setting-item">
            <label>未投喂饿了提醒 (小时)</label>
            <input type="number" v-model="settings.feedInterval" min="1" />
        </div>
        <div class="setting-item">
            <label>未互动自动游走 (分钟)</label>
            <input type="number" v-model="settings.moveInterval" min="1" />
        </div>
        <div class="timer-actions">
            <button @click="saveSettings">保存</button>
            <button @click="showSettings = false" class="cancel">取消</button>
        </div>
    </div>

    <!-- Timer Setup -->
    <div class="timer-setup" v-if="showTimer">
        <div class="selector-title">设置提醒</div>
        <div class="setting-item">
            <label>倒计时 (分钟)</label>
            <input type="number" v-model="timerDuration" min="1" max="120" />
        </div>
        <div class="setting-item">
            <label>提醒事项</label>
            <input type="text" v-model="timerMessage" placeholder="比如：喝水、休息..." />
        </div>
        <div class="timer-actions">
            <button @click="startTimer">开始</button>
            <button @click="showTimer = false" class="cancel">取消</button>
        </div>
    </div>

    <!-- Countdown Display -->
    <div class="countdown-bubble" v-if="remainingTime > 0">
        {{ formatTime(remainingTime) }}
        <span class="cancel-x" @click.stop="cancelTimer">×</span>
    </div>

    <!-- Avatar Selector -->
    <div class="avatar-selector" v-if="showAvatarSelector">
        <div class="selector-title">选择形象</div>
        <div class="avatar-grid">
            <div class="avatar-option" v-for="avatar in builtInAvatars" :key="avatar.name" @click="selectAvatar(avatar.url)">
                <img :src="avatar.url" :alt="avatar.name" />
                <span>{{ avatar.name }}</span>
            </div>
            <div class="avatar-option upload-option" @click="triggerUpload">
                <div class="upload-icon">+</div>
                <span>上传</span>
            </div>
        </div>
        <div class="close-btn" @click="showAvatarSelector = false">取消</div>
    </div>

    <div class="input-area" v-if="showInput">
      <div class="voice-toggle" @click="toggleVoiceMode" :title="isVoiceMode ? '切换键盘输入' : '切换语音输入'">
          <span v-if="isVoiceMode">⌨️</span>
          <span v-else>🎤</span>
      </div>
      
      <input v-if="!isVoiceMode" v-model="data.input" @keyup.enter="sendMessage" placeholder="说点什么..." />
      
      <button v-else class="voice-btn" :class="{ 'recording': isRecording }" 
          @mousedown="startVoiceInput" 
          @mouseup="stopVoiceInput" 
          @mouseleave="stopVoiceInput"
          @touchstart.prevent="startVoiceInput" 
          @touchend.prevent="stopVoiceInput"
      >
          {{ isRecording ? '松开发送 (正在听...)' : '按住说话' }}
      </button>
    </div>
    <input type="file" ref="fileInput" @change="handleFileUpload" accept="image/png, image/jpeg, image/gif, image/svg+xml" style="display:none">
  </div>
</template>

<style scoped>
/* ... existing styles ... */
.voice-toggle {
    cursor: pointer;
    font-size: 20px;
    padding: 5px;
    user-select: none;
    background: rgba(255, 255, 255, 0.5);
    border-radius: 50%;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.3s;
}
.voice-toggle:hover {
    background: rgba(255, 255, 255, 0.8);
}
.voice-btn {
    flex: 1;
    background-color: #42b883;
    color: white;
    border: none;
    border-radius: 4px;
    padding: 8px;
    cursor: pointer;
    font-weight: bold;
    width: 200px;
    transition: background-color 0.3s;
}
.voice-btn.recording {
    background-color: #ff4757;
    animation: pulse 1.5s infinite;
}
@keyframes pulse {
    0% { transform: scale(1); }
    50% { transform: scale(1.02); }
    100% { transform: scale(1); }
}
.companion-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  padding: 20px;
  /* Make background transparent if supported by Wails config */
}

.chat-bubble {
  background-color: rgba(240, 240, 240, 0.9);
  border-radius: 10px;
  padding: 15px;
  margin-bottom: 20px;
  position: relative;
  max-width: 300px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  color: #333;
  font-family: 'Nunito', sans-serif;
  backdrop-filter: blur(5px);
}

.chat-bubble::after {
  content: '';
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  border-width: 10px 10px 0;
  border-style: solid;
  border-color: rgba(240, 240, 240, 0.9) transparent transparent transparent;
}

.avatar img {
  max-width: 300px;
  max-height: 400px;
  width: auto;
  height: auto;
  min-width: 100px;
  min-height: 100px;
  cursor: pointer;
  transition: transform 0.2s;
  object-fit: contain;
  filter: drop-shadow(0 4px 8px rgba(0,0,0,0.1));
}

.avatar img:hover {
  transform: scale(1.05);
}

.avatar img.talking {
  animation: bounce 0.5s infinite alternate;
}

@keyframes bounce {
  from { transform: translateY(0); }
  to { transform: translateY(-10px); }
}

.input-area {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

input {
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #ccc;
  width: 200px;
  background-color: rgba(240, 240, 240, 0.9);
  backdrop-filter: blur(5px);
}

button {
  padding: 8px 16px;
  border-radius: 4px;
  border: none;
  background-color: #42b883;
  color: white;
  cursor: pointer;
  font-weight: bold;
}

button:hover {
  background-color: #33a06f;
}
</style>
