#!/bin/bash

# Claude Code omystatusline 互動式安裝程式
# Interactive installer for Claude Code omystatusline

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 安裝目錄
INSTALL_DIR="$HOME/.claude"
BINARY_NAME="statusline-go"
WRAPPER_SCRIPT="statusline-wrapper.sh"
BASH_SCRIPT="statusline.sh"

# 顯示標題
show_header() {
    clear
    echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}      ${BLUE}Claude Code omystatusline${NC} - 互動式安裝程式          ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                                ${CYAN}║${NC}"
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
    echo -ne "${CYAN}請選擇 (1-${#options[@]})${NC}: "

    read -r choice

    # 驗證輸入
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#options[@]}" ]; then
        return $((choice - 1))
    else
        return 255
    fi
}

# 檢查相依套件
check_dependencies() {
    show_header
    echo -e "${BLUE}步驟 1/5: 檢查系統相依套件${NC}"
    echo ""

    # 檢查 Go
    if ! command -v go &> /dev/null; then
        show_error "未找到 Go。請先安裝 Go 1.16 或更高版本。"
        echo ""
        show_info "安裝方式："
        echo "  macOS:   brew install go"
        echo "  Ubuntu:  sudo apt-get install golang"
        echo "  Fedora:  sudo dnf install golang"
        exit 1
    else
        local go_version=$(go version | awk '{print $3}')
        show_progress "找到 Go: $go_version"
    fi

    # 檢查 Git
    if ! command -v git &> /dev/null; then
        show_warning "未找到 Git。部分功能可能無法使用。"
    else
        show_progress "找到 Git: $(git --version | awk '{print $3}')"
    fi

    echo ""
    sleep 1
}

# 編譯二進制檔案
compile_binary() {
    show_header
    echo -e "${BLUE}步驟 2/5: 編譯 statusline 二進制檔案${NC}"
    echo ""

    show_info "正在編譯 $BINARY_NAME..."
    if go build -ldflags="-s -w" -o "$BINARY_NAME" statusline.go 2>&1 | grep -v "^#"; then
        show_progress "編譯完成"
    else
        show_error "編譯失敗"
        exit 1
    fi

    echo ""
    sleep 1
}

# 詢問音訊提醒設定
configure_audio_notifications() {
    show_header
    echo -e "${BLUE}步驟 3/5: 音訊提醒設定${NC}"
    echo ""

    show_info "音訊提醒功能可以在 Claude 完成任務時播放提示音"
    echo ""

    if ask_yes_no "是否要安裝音訊提醒功能？" "y"; then
        INSTALL_AUDIO=true
        echo ""

        # 詢問音訊類型
        ask_choice "請選擇音訊提醒方式：" \
            "🔊 使用系統預設音效（推薦）" \
            "🎵 使用自訂音訊檔案" \
            "🗣️ 使用語音播報（Text-to-Speech）"

        AUDIO_TYPE=$?

        if [ $AUDIO_TYPE -eq 255 ]; then
            show_error "無效的選擇，將使用預設音效"
            AUDIO_TYPE=0
        fi

        echo ""

        # 如果選擇自訂音訊，詢問檔案路徑
        if [ $AUDIO_TYPE -eq 1 ]; then
            echo -ne "${CYAN}請輸入自訂音訊檔案路徑 (留空使用預設)${NC}: "
            read -r CUSTOM_SOUND_PATH
            if [ -n "$CUSTOM_SOUND_PATH" ] && [ ! -f "$CUSTOM_SOUND_PATH" ]; then
                show_warning "檔案不存在，將使用預設音效"
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

# 安裝檔案
install_files() {
    show_header
    echo -e "${BLUE}步驟 4/5: 安裝檔案到 $INSTALL_DIR${NC}"
    echo ""

    # 建立目錄
    show_info "建立安裝目錄..."
    mkdir -p "$INSTALL_DIR"
    show_progress "目錄已建立"

    # 複製主要檔案
    show_info "複製主要檔案..."
    cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    cp "$WRAPPER_SCRIPT" "$INSTALL_DIR/$WRAPPER_SCRIPT"
    cp "$BASH_SCRIPT" "$INSTALL_DIR/$BASH_SCRIPT"

    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$WRAPPER_SCRIPT"
    chmod +x "$INSTALL_DIR/$BASH_SCRIPT"

    show_progress "主要檔案已安裝"

    # 安裝音訊提醒
    if [ "$INSTALL_AUDIO" = true ]; then
        show_info "設定音訊提醒功能..."

        case $AUDIO_TYPE in
            0)  # 系統預設音效
                cat > "$INSTALL_DIR/play-notification.sh" << 'EOF'
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
                show_progress "已設定系統預設音效"
                ;;

            1)  # 自訂音訊檔案
                if [ -n "$CUSTOM_SOUND_PATH" ]; then
                    # 複製自訂音訊到 .claude 目錄
                    cp "$CUSTOM_SOUND_PATH" "$INSTALL_DIR/notification-sound$(basename "$CUSTOM_SOUND_PATH" | sed 's/.*\(\.[^.]*\)$/\1/')"
                    SOUND_FILE="$INSTALL_DIR/notification-sound$(basename "$CUSTOM_SOUND_PATH" | sed 's/.*\(\.[^.]*\)$/\1/')"
                else
                    SOUND_FILE="$HOME/.claude/notification.mp3"
                fi

                cat > "$INSTALL_DIR/play-notification.sh" << EOF
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
                show_progress "已設定自訂音訊檔案"
                ;;

            2)  # 語音播報
                cat > "$INSTALL_DIR/play-notification.sh" << 'EOF'
#!/bin/bash

# 提取關鍵資訊並語音播報
INPUT=$(cat)

if echo "$INPUT" | grep -iE "error|failed" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "任務失敗，請檢查" 2>/dev/null
    elif command -v espeak &> /dev/null; then
        espeak "Task failed, please check" 2>/dev/null
    fi
elif echo "$INPUT" | grep -iE "completed|finished" > /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        say "任務完成" 2>/dev/null
    elif command -v espeak &> /dev/null; then
        espeak "Task completed" 2>/dev/null
    fi
else
    # 一般提醒音
    if [[ "$OSTYPE" == "darwin"* ]]; then
        afplay /System/Library/Sounds/Glass.aiff 2>/dev/null
    elif command -v paplay &> /dev/null; then
        paplay /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null
    else
        echo -e '\a'
    fi
fi

echo "$INPUT"
EOF
                show_progress "已設定語音播報功能"

                # 檢查 TTS 工具
                if [[ "$OSTYPE" == "darwin"* ]]; then
                    show_info "macOS 已內建 'say' 指令"
                elif ! command -v espeak &> /dev/null; then
                    show_warning "未找到 espeak。請安裝以啟用語音播報："
                    echo "  Ubuntu/Debian: sudo apt-get install espeak"
                    echo "  Fedora:        sudo dnf install espeak"
                    echo "  Arch:          sudo pacman -S espeak"
                fi
                ;;
        esac

        chmod +x "$INSTALL_DIR/play-notification.sh"
    fi

    echo ""
    show_progress "所有檔案已安裝完成"
    echo ""
    sleep 1
}

# 設定 Claude Code
configure_claude_code() {
    show_header
    echo -e "${BLUE}步驟 5/5: 設定 Claude Code${NC}"
    echo ""

    CONFIG_FILE="$INSTALL_DIR/config.json"

    # 讀取現有設定
    if [ -f "$CONFIG_FILE" ]; then
        show_info "發現現有設定檔"
        if ask_yes_no "是否要更新設定？" "y"; then
            UPDATE_CONFIG=true
        else
            UPDATE_CONFIG=false
        fi
    else
        show_info "建立新的設定檔"
        UPDATE_CONFIG=true
    fi

    if [ "$UPDATE_CONFIG" = true ]; then
        # 備份現有設定
        if [ -f "$CONFIG_FILE" ]; then
            cp "$CONFIG_FILE" "$CONFIG_FILE.backup.$(date +%Y%m%d%H%M%S)"
            show_progress "已備份現有設定"
        fi

        # 建立或更新設定
        if [ "$INSTALL_AUDIO" = true ]; then
            # 包含音訊提醒的設定
            cat > "$CONFIG_FILE" << EOF
{
  "statusLineCommand": "$INSTALL_DIR/$WRAPPER_SCRIPT",
  "hooks": {
    "assistantMessageEnd": "$INSTALL_DIR/play-notification.sh"
  }
}
EOF
            show_progress "已設定狀態列與音訊提醒"
        else
            # 僅狀態列設定
            cat > "$CONFIG_FILE" << EOF
{
  "statusLineCommand": "$INSTALL_DIR/$WRAPPER_SCRIPT"
}
EOF
            show_progress "已設定狀態列"
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
    echo -e "${GREEN}║${NC}                     ${BLUE}✓ 安裝完成！${NC}                            ${GREEN}║${NC}"
    echo -e "${GREEN}║${NC}                                                                ${GREEN}║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    echo -e "${BLUE}已安裝的檔案：${NC}"
    echo "  ✓ $INSTALL_DIR/$BINARY_NAME"
    echo "  ✓ $INSTALL_DIR/$WRAPPER_SCRIPT"
    echo "  ✓ $INSTALL_DIR/$BASH_SCRIPT"

    if [ "$INSTALL_AUDIO" = true ]; then
        echo "  ✓ $INSTALL_DIR/play-notification.sh"
        case $AUDIO_TYPE in
            0) echo "     └─ 使用系統預設音效" ;;
            1) echo "     └─ 使用自訂音訊檔案" ;;
            2) echo "     └─ 使用語音播報（TTS）" ;;
        esac
    fi

    echo ""
    echo -e "${BLUE}設定檔位置：${NC}"
    echo "  ✓ $INSTALL_DIR/config.json"

    echo ""
    echo -e "${YELLOW}下一步：${NC}"
    echo "  1. 重新啟動 Claude Code 或開始新的對話"
    echo "  2. 你應該會看到新的狀態列顯示"

    if [ "$INSTALL_AUDIO" = true ]; then
        echo "  3. 當 Claude 完成回覆時會播放提示音"
        echo ""
        echo -e "${CYAN}測試音訊提醒：${NC}"
        echo "  $INSTALL_DIR/play-notification.sh"
    fi

    echo ""
    echo -e "${CYAN}更多資訊：${NC}"
    echo "  - README: https://github.com/howie/claude-code-omystatusline"
    echo "  - 音訊提醒文件: docs/features/audio-notifications/README.md"

    echo ""
    echo -e "${GREEN}感謝使用 Claude Code omystatusline！${NC}"
    echo ""
}

# 主程式流程
main() {
    # 檢查是否在專案目錄
    if [ ! -f "statusline.go" ]; then
        show_error "請在 claude-code-omystatusline 專案目錄中執行此腳本"
        exit 1
    fi

    # 執行安裝步驟
    check_dependencies
    compile_binary
    configure_audio_notifications
    install_files
    configure_claude_code
    show_summary

    # 清理暫存檔案
    rm -f "$BINARY_NAME"
}

# 執行主程式
main
