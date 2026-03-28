#!/bin/bash

# Claude Code omystatusline Interactive Installer
# Claude Code omystatusline 互動式安裝程式

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 安裝目錄
INSTALL_DIR="$HOME/.claude/omystatusline"
CLAUDE_DIR="$HOME/.claude"
OUTPUT_DIR="output"
BINARY_NAME="statusline-go"
WRAPPER_SCRIPT="statusline-wrapper.sh"
BASH_SCRIPT="statusline.sh"

# 預設語系：英文
LANG_CHOICE="en"

# ============================================================================
# 多語系訊息定義
# ============================================================================

# 訊息函式：根據語系返回對應文字
msg() {
    local key="$1"

    case "$LANG_CHOICE" in
        zh)
            case "$key" in
                # 標題
                "title") echo "Claude Code omystatusline - 互動式安裝程式" ;;

                # 語系選擇
                "lang_prompt") echo "請選擇語言 / Choose Language" ;;
                "lang_en") echo "English" ;;
                "lang_zh") echo "繁體中文" ;;
                "invalid_choice") echo "無效的選擇，使用預設英文" ;;

                # 步驟標題
                "step_check_deps") echo "步驟 1/5: 檢查系統相依套件" ;;
                "step_compile") echo "步驟 2/5: 編譯 statusline 二進制檔案" ;;
                "step_audio") echo "步驟 3/5: 音訊提醒設定" ;;
                "step_install") echo "步驟 4/5: 安裝檔案到 $INSTALL_DIR" ;;
                "step_config") echo "步驟 5/5: 設定 Claude Code" ;;

                # 相依性檢查
                "go_not_found") echo "未找到 Go。請先安裝 Go 1.16 或更高版本。" ;;
                "install_methods") echo "安裝方式：" ;;
                "found_go") echo "找到 Go:" ;;
                "git_not_found") echo "未找到 Git。部分功能可能無法使用。" ;;
                "found_git") echo "找到 Git:" ;;

                # 編譯
                "compiling") echo "正在編譯 $BINARY_NAME..." ;;
                "compile_success") echo "編譯完成" ;;
                "compile_failed") echo "編譯失敗" ;;

                # 音訊提醒
                "audio_desc") echo "音訊提醒功能可以在 Claude 完成任務時播放提示音" ;;
                "audio_install_q") echo "是否要安裝音訊提醒功能？" ;;
                "audio_mode_q") echo "請選擇音訊提醒方式：" ;;
                "audio_system") echo "🔊 使用系統預設音效（推薦）" ;;
                "audio_custom") echo "🎵 使用自訂音訊檔案" ;;
                "audio_tts") echo "🗣️ 使用語音播報（Text-to-Speech）" ;;
                "choose_1_3") echo "請選擇 (1-3)" ;;
                "invalid_using_default") echo "無效的選擇，將使用預設音效" ;;
                "custom_path_prompt") echo "請輸入自訂音訊檔案路徑 (留空使用預設)" ;;
                "file_not_exist") echo "檔案不存在，將使用預設音效" ;;
                "test_tts_q") echo "是否要測試 TTS 語音播報？" ;;
                "test_tts_success") echo "測試成功訊息播報" ;;
                "test_tts_error") echo "測試錯誤訊息播報" ;;
                "test_tts_general") echo "測試一般提示音" ;;
                "testing_in_progress") echo "正在播放測試語音..." ;;

                # 安裝檔案
                "creating_dir") echo "建立安裝目錄..." ;;
                "dir_created") echo "目錄已建立" ;;
                "copying_files") echo "複製主要檔案..." ;;
                "files_installed") echo "主要檔案已安裝" ;;
                "configuring_audio") echo "設定音訊提醒功能..." ;;
                "audio_system_done") echo "已設定系統預設音效" ;;
                "audio_custom_done") echo "已設定自訂音訊檔案" ;;
                "audio_tts_done") echo "已設定語音播報功能" ;;
                "tts_builtin") echo "macOS 已內建 'say' 指令" ;;
                "tts_not_found") echo "未找到 espeak。請安裝以啟用語音播報：" ;;
                "install_complete") echo "所有檔案已安裝完成" ;;

                # 設定
                "found_config") echo "發現現有設定檔" ;;
                "update_config_q") echo "是否要更新設定？" ;;
                "creating_config") echo "建立新的設定檔" ;;
                "backup_config") echo "已備份現有設定" ;;
                "config_statusline_audio") echo "已設定狀態列與音訊提醒" ;;
                "config_statusline") echo "已設定狀態列" ;;

                # 完成摘要
                "install_success") echo "✓ 安裝完成！" ;;
                "installed_files") echo "已安裝的檔案：" ;;
                "using_system_sound") echo "└─ 使用系統預設音效" ;;
                "using_custom_sound") echo "└─ 使用自訂音訊檔案" ;;
                "using_tts") echo "└─ 使用語音播報（TTS）" ;;
                "config_location") echo "設定檔位置：" ;;
                "next_steps") echo "下一步：" ;;
                "next_1") echo "1. 重新啟動 Claude Code 或開始新的對話" ;;
                "next_2") echo "2. 你應該會看到新的狀態列顯示" ;;
                "next_3") echo "3. 當 Claude 完成回覆時會播放提示音" ;;
                "test_audio") echo "測試音訊提醒：" ;;
                "more_info") echo "更多資訊：" ;;
                "readme_link") echo "- README: https://github.com/howie/claude-code-omystatusline" ;;
                "audio_doc") echo "- 音訊提醒文件: docs/features/audio-notifications/README.md" ;;
                "thanks") echo "感謝使用 Claude Code omystatusline！" ;;

                *) echo "$key" ;;
            esac
            ;;

        *)  # 預設英文
            case "$key" in
                # Title
                "title") echo "Claude Code omystatusline - Interactive Installer" ;;

                # Language selection
                "lang_prompt") echo "Choose Language / 請選擇語言" ;;
                "lang_en") echo "English" ;;
                "lang_zh") echo "繁體中文 (Traditional Chinese)" ;;
                "invalid_choice") echo "Invalid choice, using default English" ;;

                # Step titles
                "step_check_deps") echo "Step 1/5: Checking System Dependencies" ;;
                "step_compile") echo "Step 2/5: Compiling statusline binary" ;;
                "step_audio") echo "Step 3/5: Audio Notification Configuration" ;;
                "step_install") echo "Step 4/5: Installing files to $INSTALL_DIR" ;;
                "step_config") echo "Step 5/5: Configuring Claude Code" ;;

                # Dependency check
                "go_not_found") echo "Go not found. Please install Go 1.16 or higher." ;;
                "install_methods") echo "Installation methods:" ;;
                "found_go") echo "Found Go:" ;;
                "git_not_found") echo "Git not found. Some features may not work." ;;
                "found_git") echo "Found Git:" ;;

                # Compilation
                "compiling") echo "Compiling $BINARY_NAME..." ;;
                "compile_success") echo "Compilation completed" ;;
                "compile_failed") echo "Compilation failed" ;;

                # Audio notifications
                "audio_desc") echo "Audio notifications can play sounds when Claude completes tasks" ;;
                "audio_install_q") echo "Would you like to install audio notification features?" ;;
                "audio_mode_q") echo "Please select audio notification mode:" ;;
                "audio_system") echo "🔊 Use system default sounds (recommended)" ;;
                "audio_custom") echo "🎵 Use custom audio file" ;;
                "audio_tts") echo "🗣️ Use text-to-speech (TTS)" ;;
                "choose_1_3") echo "Choose (1-3)" ;;
                "invalid_using_default") echo "Invalid choice, will use default sound" ;;
                "custom_path_prompt") echo "Enter custom audio file path (leave empty for default)" ;;
                "file_not_exist") echo "File does not exist, will use default sound" ;;
                "test_tts_q") echo "Would you like to test TTS voice output?" ;;
                "test_tts_success") echo "Testing success message" ;;
                "test_tts_error") echo "Testing error message" ;;
                "test_tts_general") echo "Testing general notification" ;;
                "testing_in_progress") echo "Playing test audio..." ;;

                # Installing files
                "creating_dir") echo "Creating installation directory..." ;;
                "dir_created") echo "Directory created" ;;
                "copying_files") echo "Copying main files..." ;;
                "files_installed") echo "Main files installed" ;;
                "configuring_audio") echo "Configuring audio notifications..." ;;
                "audio_system_done") echo "System default sound configured" ;;
                "audio_custom_done") echo "Custom audio file configured" ;;
                "audio_tts_done") echo "Text-to-speech configured" ;;
                "tts_builtin") echo "macOS has built-in 'say' command" ;;
                "tts_not_found") echo "espeak not found. Install to enable TTS:" ;;
                "install_complete") echo "All files installed successfully" ;;

                # Configuration
                "found_config") echo "Found existing configuration" ;;
                "update_config_q") echo "Would you like to update the configuration?" ;;
                "creating_config") echo "Creating new configuration file" ;;
                "backup_config") echo "Backed up existing configuration" ;;
                "config_statusline_audio") echo "Configured status line and audio notifications" ;;
                "config_statusline") echo "Configured status line" ;;

                # Completion summary
                "install_success") echo "✓ Installation Complete!" ;;
                "installed_files") echo "Installed files:" ;;
                "using_system_sound") echo "└─ Using system default sound" ;;
                "using_custom_sound") echo "└─ Using custom audio file" ;;
                "using_tts") echo "└─ Using text-to-speech (TTS)" ;;
                "config_location") echo "Configuration location:" ;;
                "next_steps") echo "Next steps:" ;;
                "next_1") echo "1. Restart Claude Code or start a new conversation" ;;
                "next_2") echo "2. You should see the new status line display" ;;
                "next_3") echo "3. Audio will play when Claude completes a response" ;;
                "test_audio") echo "Test audio notification:" ;;
                "more_info") echo "More information:" ;;
                "readme_link") echo "- README: https://github.com/howie/claude-code-omystatusline" ;;
                "audio_doc") echo "- Audio notifications doc: docs/features/audio-notifications/README.md" ;;
                "thanks") echo "Thank you for using Claude Code omystatusline!" ;;

                *) echo "$key" ;;
            esac
            ;;
    esac
}

# ============================================================================
# 顯示函式
# ============================================================================

# 顯示標題
show_header() {
    clear
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}      ${BLUE}$(msg "title")${NC}"
    local padding=$((64 - ${#title_text}))
    printf "${CYAN}║${NC}%*s${CYAN}║${NC}\n" $padding ""
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# 顯示進度
show_progress() {
    echo -e "${GREEN}✓${NC} $1"
}

# 顯示錯誤
show_error() {
    echo -e "${RED}✗${NC} $1"
}

# 顯示警告
show_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# 顯示資訊
show_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# 詢問是非題
ask_yes_no() {
    local question="$1"
    local default="$2"
    local response

    if [ "$default" = "y" ]; then
        echo -ne "${CYAN}?${NC} $question ${GREEN}[Y/n]${NC}: "
    else
        echo -ne "${CYAN}?${NC} $question ${YELLOW}[y/N]${NC}: "
    fi

    read -r response
    response=$(echo "$response" | tr '[:upper:]' '[:lower:]')

    if [ -z "$response" ]; then
        response="$default"
    fi

    [ "$response" = "y" ] || [ "$response" = "yes" ]
}

# 選擇選項
ask_choice() {
    local question="$1"
    shift
    local options=("$@")
    local choice

    echo -e "${CYAN}?${NC} $question"
    for i in "${!options[@]}"; do
        echo "  $((i+1)). ${options[$i]}"
    done
    echo -ne "${CYAN}$(msg "choose_1_3")${NC}: "

    read -r choice

    # 驗證輸入
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#options[@]}" ]; then
        return $((choice - 1))
    else
        return 255
    fi
}

# ============================================================================
# 語系選擇
# ============================================================================

choose_language() {
    clear
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}         ${BLUE}Claude Code omystatusline${NC}                         ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    echo -e "${CYAN}?${NC} $(msg "lang_prompt")"
    echo "  1. $(msg "lang_en")"
    echo "  2. $(msg "lang_zh")"
    echo -ne "${CYAN}Choose / 選擇 (1-2) [1]${NC}: "

    read -r lang_input

    case "$lang_input" in
        2)
            LANG_CHOICE="zh"
            ;;
        1|"")
            LANG_CHOICE="en"
            ;;
        *)
            LANG_CHOICE="en"
            show_warning "$(msg "invalid_choice")"
            sleep 1
            ;;
    esac

    echo ""
    sleep 1
}

# ============================================================================
# 安裝步驟
# ============================================================================

# 檢查相依套件
check_dependencies() {
    show_header
    echo -e "${BLUE}$(msg "step_check_deps")${NC}"
    echo ""

    # 檢查 Go
    if ! command -v go &> /dev/null; then
        show_error "$(msg "go_not_found")"
        echo ""
        show_info "$(msg "install_methods")"
        echo "  macOS:   brew install go"
        echo "  Ubuntu:  sudo apt-get install golang"
        echo "  Fedora:  sudo dnf install golang"
        exit 1
    else
        local go_version=$(go version | awk '{print $3}')
        show_progress "$(msg "found_go") $go_version"
    fi

    # 檢查 Git
    if ! command -v git &> /dev/null; then
        show_warning "$(msg "git_not_found")"
    else
        show_progress "$(msg "found_git") $(git --version | awk '{print $3}')"
    fi

    echo ""
    sleep 1
}

# 編譯二進制檔案
compile_binary() {
    show_header
    echo -e "${BLUE}$(msg "step_compile")${NC}"
    echo ""

    show_info "$(msg "compiling")"
    # 編譯新的 cmd/statusline 目錄下的主程式
    mkdir -p "$OUTPUT_DIR"
    if go build -ldflags="-s -w" -o "$OUTPUT_DIR/$BINARY_NAME" ./cmd/statusline 2>&1; then
        show_progress "$(msg "compile_success")"
    else
        show_error "$(msg "compile_failed")"
        exit 1
    fi

    # 編譯 voice-reminder
    show_info "正在編譯 voice-reminder..."
    if go build -ldflags="-s -w" -o "$OUTPUT_DIR/voice-reminder" ./cmd/voice-reminder 2>&1; then
        show_progress "voice-reminder 編譯完成"
    else
        show_error "voice-reminder 編譯失敗"
        exit 1
    fi

    echo ""
    sleep 1
}

# 詢問音訊提醒設定
configure_audio_notifications() {
    show_header
    echo -e "${BLUE}$(msg "step_audio")${NC}"
    echo ""

    show_info "$(msg "audio_desc")"
    echo ""

    if ask_yes_no "$(msg "audio_install_q")" "y"; then
        INSTALL_AUDIO=true
        echo ""

        # 詢問音訊類型
        set +e  # 暫時禁用 set -e，因為 ask_choice 使用返回值傳遞選擇結果
        ask_choice "$(msg "audio_mode_q")" \
            "$(msg "audio_system")" \
            "$(msg "audio_custom")" \
            "$(msg "audio_tts")"

        AUDIO_TYPE=$?
        set -e  # 重新啟用 set -e

        if [ $AUDIO_TYPE -eq 255 ]; then
            show_error "$(msg "invalid_using_default")"
            AUDIO_TYPE=0
        fi

        echo ""

        # 如果選擇自訂音訊，詢問檔案路徑
        if [ $AUDIO_TYPE -eq 1 ]; then
            echo -ne "${CYAN}$(msg "custom_path_prompt")${NC}: "
            read -r CUSTOM_SOUND_PATH
            if [ -n "$CUSTOM_SOUND_PATH" ] && [ ! -f "$CUSTOM_SOUND_PATH" ]; then
                show_warning "$(msg "file_not_exist")"
                AUDIO_TYPE=0
                CUSTOM_SOUND_PATH=""
            fi
        fi
    else
        INSTALL_AUDIO=false
    fi

    echo ""
    sleep 1
}

# 測試 TTS 語音播報
test_tts_notification() {
    local test_script="$1"

    show_header
    echo -e "${BLUE}$(msg "step_audio") - TTS Test${NC}"
    echo ""

    if ask_yes_no "$(msg "test_tts_q")" "y"; then
        echo ""

        # 測試 1: 成功訊息
        show_info "$(msg "test_tts_success")"
        echo "Task completed successfully" | "$test_script"
        sleep 2

        # 測試 2: 錯誤訊息
        show_info "$(msg "test_tts_error")"
        echo "Error: something failed" | "$test_script"
        sleep 2

        # 測試 3: 一般通知
        show_info "$(msg "test_tts_general")"
        echo "General notification message" | "$test_script"
        sleep 2

        echo ""
        show_progress "$(msg "testing_in_progress")"
    fi

    echo ""
    sleep 1
}

# 安裝檔案
install_files() {
    show_header
    echo -e "${BLUE}$(msg "step_install")${NC}"
    echo ""

    # 建立目錄結構
    show_info "$(msg "creating_dir")"
    mkdir -p "$INSTALL_DIR/bin"
    mkdir -p "$INSTALL_DIR/scripts"
    mkdir -p "$INSTALL_DIR/plugins/voice-reminder/bin"
    mkdir -p "$INSTALL_DIR/plugins/voice-reminder/config"
    mkdir -p "$INSTALL_DIR/plugins/voice-reminder/scripts"
    mkdir -p "$INSTALL_DIR/plugins/voice-reminder/data"
    mkdir -p "$INSTALL_DIR/plugins/voice-reminder/commands"
    mkdir -p "$CLAUDE_DIR/commands"
    show_progress "$(msg "dir_created")"

    # 複製 statusline 主要檔案
    show_info "$(msg "copying_files")"
    cp "$OUTPUT_DIR/$BINARY_NAME" "$INSTALL_DIR/bin/$BINARY_NAME"
    cp "scripts/$WRAPPER_SCRIPT" "$INSTALL_DIR/bin/$WRAPPER_SCRIPT"
    cp "scripts/$BASH_SCRIPT" "$INSTALL_DIR/scripts/$BASH_SCRIPT"

    chmod +x "$INSTALL_DIR/bin/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/bin/$WRAPPER_SCRIPT"
    chmod +x "$INSTALL_DIR/scripts/$BASH_SCRIPT"

    show_progress "$(msg "files_installed")"

    # 安裝 voice-reminder plugin
    show_info "正在安裝 voice-reminder plugin..."
    cp "$OUTPUT_DIR/voice-reminder" "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder"
    chmod +x "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder"

    # 複製配置檔案（如果不存在）
    if [ ! -f "$INSTALL_DIR/plugins/voice-reminder/config/voice-reminder-config.json" ]; then
        cp "configs/voice-reminder-config.json" "$INSTALL_DIR/plugins/voice-reminder/config/voice-reminder-config.json"
    fi

    # 複製 voice-reminder 腳本
    cp "scripts/toggle-voice-reminder.sh" "$INSTALL_DIR/plugins/voice-reminder/scripts/"
    cp "scripts/test-voice-reminder.sh" "$INSTALL_DIR/plugins/voice-reminder/scripts/"
    chmod +x "$INSTALL_DIR/plugins/voice-reminder/scripts/toggle-voice-reminder.sh"
    chmod +x "$INSTALL_DIR/plugins/voice-reminder/scripts/test-voice-reminder.sh"

    # 複製 slash commands 並建立符號連結
    cp commands/voice-reminder-*.md "$INSTALL_DIR/plugins/voice-reminder/commands/"

    # 建立符號連結到 ~/.claude/commands/
    for cmd_file in commands/voice-reminder-*.md; do
        cmd_name=$(basename "$cmd_file")
        ln -sf "$INSTALL_DIR/plugins/voice-reminder/commands/$cmd_name" "$CLAUDE_DIR/commands/$cmd_name"
    done

    # 預設啟用 voice-reminder
    echo "true" > "$INSTALL_DIR/plugins/voice-reminder/data/voice-reminder-enabled"

    show_progress "voice-reminder plugin 已安裝"

    # 安裝音訊提醒
    if [ "$INSTALL_AUDIO" = true ]; then
        show_info "$(msg "configuring_audio")"

        case $AUDIO_TYPE in
            0)  # 系統預設音效
                cat > "$INSTALL_DIR/scripts/play-notification.sh" << 'EOF'
#!/bin/bash

# 根據作業系統選擇音訊播放工具
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    afplay /System/Library/Sounds/Glass.aiff 2>/dev/null
elif command -v paplay &> /dev/null; then
    # Linux with PulseAudio
    paplay /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null
elif command -v aplay &> /dev/null; then
    # Linux with ALSA
    aplay /usr/share/sounds/alsa/Front_Center.wav 2>/dev/null
elif command -v beep &> /dev/null; then
    # 使用系統蜂鳴器
    beep -f 800 -l 200 2>/dev/null
else
    # 使用終端機鈴聲作為備援方案
    echo -e '\a'
fi
EOF
                show_progress "$(msg "audio_system_done")"
                ;;

            1)  # 自訂音訊檔案
                if [ -n "$CUSTOM_SOUND_PATH" ]; then
                    # 複製自訂音訊到 omystatusline 目錄
                    cp "$CUSTOM_SOUND_PATH" "$INSTALL_DIR/scripts/notification-sound$(basename "$CUSTOM_SOUND_PATH" | sed 's/.*\(\.[^.]*\)$/\1/')"
                    SOUND_FILE="$INSTALL_DIR/scripts/notification-sound$(basename "$CUSTOM_SOUND_PATH" | sed 's/.*\(\.[^.]*\)$/\1/')"
                else
                    SOUND_FILE="$INSTALL_DIR/scripts/notification.mp3"
                fi

                cat > "$INSTALL_DIR/scripts/play-notification.sh" << EOF
#!/bin/bash

SOUND_FILE="$SOUND_FILE"

if [[ "$OSTYPE" == "darwin"* ]]; then
    afplay "\$SOUND_FILE" 2>/dev/null
elif command -v ffplay &> /dev/null; then
    ffplay -nodisp -autoexit "\$SOUND_FILE" 2>/dev/null
elif command -v mpg123 &> /dev/null; then
    mpg123 -q "\$SOUND_FILE" 2>/dev/null
elif command -v paplay &> /dev/null && command -v ffmpeg &> /dev/null; then
    ffmpeg -i "\$SOUND_FILE" -f wav - 2>/dev/null | paplay
else
    echo -e '\a'
fi
EOF
                show_progress "$(msg "audio_custom_done")"
                ;;

            2)  # 語音播報
                cat > "$INSTALL_DIR/scripts/play-notification.sh" << EOF
#!/bin/bash

# TTS Language Setting (set during installation)
TTS_LANG="$LANG_CHOICE"

# 提取關鍵資訊並語音播報
INPUT=\$(cat)

# 從 JSON 提取 message 欄位
MESSAGE=\$(echo "\$INPUT" | jq -r '.message // ""' 2>/dev/null)

# 如果 jq 不可用或失敗，使用 grep/sed 作為備援
if [ -z "\$MESSAGE" ] || [ "\$MESSAGE" = "null" ]; then
    MESSAGE=\$(echo "\$INPUT" | grep -o '"message"[[:space:]]*:[[:space:]]*"[^"]*"' | sed 's/"message"[[:space:]]*:[[:space:]]*"\(.*\)"/\1/' 2>/dev/null)
fi

# 如果還是提取不到，使用整個輸入
if [ -z "\$MESSAGE" ]; then
    MESSAGE="\$INPUT"
fi

# 檢查是否需要使用者確認（問號 OR 關鍵字）
# 支援中文問號（？）和英文問號（?）
if echo "\$MESSAGE" | grep -E "[?？]" > /dev/null || \
   echo "\$MESSAGE" | grep -iE "permission|confirm|approve" > /dev/null; then
    if [[ "\$OSTYPE" == "darwin"* ]]; then
        if [ "\$TTS_LANG" = "zh" ]; then
            say "Claude 需要您的確認" 2>/dev/null
        else
            say "Claude needs your confirmation" 2>/dev/null
        fi
    elif command -v espeak &> /dev/null; then
        espeak "Claude needs your confirmation" 2>/dev/null
    fi

# 檢查錯誤
elif echo "\$MESSAGE" | grep -iE "error|failed" > /dev/null; then
    if [[ "\$OSTYPE" == "darwin"* ]]; then
        if [ "\$TTS_LANG" = "zh" ]; then
            say "任務失敗，請檢查" 2>/dev/null
        else
            say "Task failed, please check" 2>/dev/null
        fi
    elif command -v espeak &> /dev/null; then
        espeak "Task failed, please check" 2>/dev/null
    fi

# 檢查完成
elif echo "\$MESSAGE" | grep -iE "completed|finished" > /dev/null; then
    if [[ "\$OSTYPE" == "darwin"* ]]; then
        if [ "\$TTS_LANG" = "zh" ]; then
            say "任務完成" 2>/dev/null
        else
            say "Task completed" 2>/dev/null
        fi
    elif command -v espeak &> /dev/null; then
        espeak "Task completed" 2>/dev/null
    fi

# 其他情況播放音效
else
    if [[ "\$OSTYPE" == "darwin"* ]]; then
        afplay /System/Library/Sounds/Glass.aiff 2>/dev/null
    elif command -v paplay &> /dev/null; then
        paplay /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null
    else
        echo -e '\a'
    fi
fi

echo "\$INPUT"
EOF
                show_progress "$(msg "audio_tts_done")"

                # 檢查 TTS 工具
                if [[ "$OSTYPE" == "darwin"* ]]; then
                    show_info "$(msg "tts_builtin")"
                elif ! command -v espeak &> /dev/null; then
                    show_warning "$(msg "tts_not_found")"
                    echo "  Ubuntu/Debian: sudo apt-get install espeak"
                    echo "  Fedora:        sudo dnf install espeak"
                    echo "  Arch:          sudo pacman -S espeak"
                fi
                ;;
        esac

        chmod +x "$INSTALL_DIR/scripts/play-notification.sh"

        # 如果是 TTS 模式，提供測試選項
        if [ $AUDIO_TYPE -eq 2 ]; then
            echo ""
            test_tts_notification "$INSTALL_DIR/scripts/play-notification.sh"
        fi
    fi

    echo ""
    show_progress "$(msg "install_complete")"
    echo ""
    sleep 1
}

# 設定 Claude Code
configure_claude_code() {
    show_header
    echo -e "${BLUE}$(msg "step_config")${NC}"
    echo ""

    # 使用 settings.json（Claude Code v2.0.25+ 的新配置文件）
    CONFIG_FILE="$CLAUDE_DIR/settings.json"

    # 讀取現有設定
    if [ -f "$CONFIG_FILE" ]; then
        show_info "$(msg "found_config")"
        if ask_yes_no "$(msg "update_config_q")" "y"; then
            UPDATE_CONFIG=true
        else
            UPDATE_CONFIG=false
        fi
    else
        show_info "$(msg "creating_config")"
        UPDATE_CONFIG=true
    fi

    if [ "$UPDATE_CONFIG" = true ]; then
        # 備份現有設定
        if [ -f "$CONFIG_FILE" ]; then
            cp "$CONFIG_FILE" "$CONFIG_FILE.backup.$(date +%Y%m%d%H%M%S)"
            show_progress "$(msg "backup_config")"
        fi

        # 讀取現有配置以保留其他設定（如 mcpServers, alwaysThinkingEnabled 等）
        EXISTING_CONFIG=""
        if ls "$CONFIG_FILE.backup."* 1> /dev/null 2>&1; then
            LATEST_BACKUP=$(ls -t "$CONFIG_FILE.backup."* 2>/dev/null | head -1)
            if [ -n "$LATEST_BACKUP" ]; then
                EXISTING_CONFIG=$(cat "$LATEST_BACKUP")
            fi
        fi

        # 建立或更新設定（使用新版格式和 voice-reminder 系統）
        if [ "$INSTALL_AUDIO" = true ]; then
            # 包含 voice-reminder 的設定（支援四個 hook 事件：Notification, Stop, SubagentStop, PreToolUse）
            if [ -n "$EXISTING_CONFIG" ] && command -v jq &> /dev/null; then
                # 使用 jq 合併配置（如果有的話）
                echo "$EXISTING_CONFIG" | jq \
                    --arg wrapper "$INSTALL_DIR/bin/$WRAPPER_SCRIPT" \
                    --arg reminder "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder" \
                    '.statusLine = {"type": "command", "command": $wrapper, "padding": 0} |
                     .hooks = {
                       "Notification": [{"matcher": "", "hooks": [{"type": "command", "command": $reminder}]}],
                       "Stop": [{"matcher": "", "hooks": [{"type": "command", "command": $reminder}]}],
                       "SubagentStop": [{"matcher": "", "hooks": [{"type": "command", "command": $reminder}]}],
                       "PreToolUse": [{"matcher": "", "hooks": [{"type": "command", "command": $reminder}]}]
                     }' > "$CONFIG_FILE"
            else
                # 沒有 jq 或沒有現有配置，建立基本配置
                cat > "$CONFIG_FILE" << EOF
{
  "statusLine": {
    "type": "command",
    "command": "$INSTALL_DIR/bin/$WRAPPER_SCRIPT",
    "padding": 0
  },
  "hooks": {
    "Notification": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder"
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder"
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "$INSTALL_DIR/plugins/voice-reminder/bin/voice-reminder"
          }
        ]
      }
    ]
  }
}
EOF
            fi
            show_progress "$(msg "config_statusline_audio")"
        else
            # 僅狀態列設定
            if [ -n "$EXISTING_CONFIG" ] && command -v jq &> /dev/null; then
                echo "$EXISTING_CONFIG" | jq \
                    --arg wrapper "$INSTALL_DIR/bin/$WRAPPER_SCRIPT" \
                    '.statusLine = {"type": "command", "command": $wrapper, "padding": 0}' > "$CONFIG_FILE"
            else
                cat > "$CONFIG_FILE" << EOF
{
  "statusLine": {
    "type": "command",
    "command": "$INSTALL_DIR/bin/$WRAPPER_SCRIPT",
    "padding": 0
  }
}
EOF
            fi
            show_progress "$(msg "config_statusline")"
        fi
    fi

    echo ""
    sleep 1
}

# 顯示安裝摘要
show_summary() {
    show_header
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}                     ${BLUE}$(msg "install_success")${NC}                            ${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    echo -e "${BLUE}$(msg "installed_files")${NC}"
    echo "  ✓ $INSTALL_DIR/bin/$BINARY_NAME"
    echo "  ✓ $INSTALL_DIR/bin/$WRAPPER_SCRIPT"
    echo "  ✓ $INSTALL_DIR/scripts/$BASH_SCRIPT"
    echo "  ✓ $INSTALL_DIR/plugins/voice-reminder/"

    if [ "$INSTALL_AUDIO" = true ]; then
        echo "  ✓ $INSTALL_DIR/scripts/play-notification.sh"
        case $AUDIO_TYPE in
            0) echo "     $(msg "using_system_sound")" ;;
            1) echo "     $(msg "using_custom_sound")" ;;
            2) echo "     $(msg "using_tts")" ;;
        esac
    fi

    echo ""
    echo -e "${BLUE}$(msg "config_location")${NC}"
    echo "  ✓ $CLAUDE_DIR/settings.json"

    echo ""
    echo -e "${YELLOW}$(msg "next_steps")${NC}"
    echo "  $(msg "next_1")"
    echo "  $(msg "next_2")"

    if [ "$INSTALL_AUDIO" = true ]; then
        echo "  $(msg "next_3")"
        echo ""
        echo -e "${CYAN}$(msg "test_audio")${NC}"
        echo "  $INSTALL_DIR/scripts/play-notification.sh"
    fi

    echo ""
    echo -e "${CYAN}$(msg "more_info")${NC}"
    echo "  $(msg "readme_link")"
    echo "  $(msg "audio_doc")"

    echo ""
    echo -e "${GREEN}$(msg "thanks")${NC}"
    echo ""
}

# ============================================================================
# 主程式流程
# ============================================================================

main() {
    # 檢查是否在專案目錄 (檢查新的目錄結構)
    if [ ! -d "cmd/statusline" ] || [ ! -f "go.mod" ]; then
        show_error "Please run this script in the claude-code-omystatusline project directory"
        show_error "請在 claude-code-omystatusline 專案目錄中執行此腳本"
        exit 1
    fi

    # 選擇語系
    choose_language

    # 執行安裝步驟
    check_dependencies
    compile_binary
    configure_audio_notifications
    install_files
    configure_claude_code
    show_summary

    # 清理暫存檔案
    rm -rf "$OUTPUT_DIR"
}

# 執行主程式
main
