#!/bin/bash

# Claude 時間統計工具
# 功能：統計和管理 Claude Code session 使用時間
#
# 用法：
#   claude-stats.sh          # 今日統計
#   claude-stats.sh 2025-08-08  # 指定日期統計
#   claude-stats.sh week      # 本週統計
#   claude-stats.sh month     # 本月統計
#   claude-stats.sh all       # 所有歷史統計
#   claude-stats.sh archive   # 歸檔舊 session 檔案
#
# 說明：
#   - 統計功能會同時檢查 sessions 和 archive 目錄
#   - 歸檔功能會將非今日的 session 檔案移動到按日期分類的 archive 目錄
#   - 支援中文星期顯示和時間格式化

TRACKER_DIR="$HOME/.claude/session-tracker"
SESSIONS_DIR="$TRACKER_DIR/sessions"
ARCHIVE_DIR="$TRACKER_DIR/archive"

# 將英文星期轉換為中文
get_chinese_weekday() {
    local date_str=$1
    local weekday

    # 根據平台使用不同的 date 命令
    if date --version >/dev/null 2>&1; then
        # GNU date (Linux)
        weekday=$(date -d "$date_str" +%a 2>/dev/null)
    else
        # BSD date (macOS)
        weekday=$(date -j -f "%Y-%m-%d" "$date_str" +%a 2>/dev/null)
    fi

    case "$weekday" in
        Mon) echo "(一)" ;;
        Tue) echo "(二)" ;;
        Wed) echo "(三)" ;;
        Thu) echo "(四)" ;;
        Fri) echo "(五)" ;;
        Sat) echo "(六)" ;;
        Sun) echo "(日)" ;;
        *) echo "" ;;
    esac
}

# 計算指定日期的總時數
calculate_date_total() {
    local date=$1
    local total_seconds=0
    local session_count=0

    # 檢查活躍目錄（移除日期限制）
    for session_file in "$SESSIONS_DIR"/*.json; do
        if [ -f "$session_file" ]; then
            local session_date=$(jq -r '.date' "$session_file" 2>/dev/null)
            if [ "$session_date" = "$date" ]; then
                local session_seconds=$(jq '.total_seconds // 0' "$session_file")
                total_seconds=$((total_seconds + session_seconds))
                session_count=$((session_count + 1))
            fi
        fi
    done

    # 檢查歸檔目錄
    if [ -d "$ARCHIVE_DIR/$date" ]; then
        for session_file in "$ARCHIVE_DIR/$date"/*.json; do
            if [ -f "$session_file" ]; then
                local session_seconds=$(jq '.total_seconds // 0' "$session_file")
                total_seconds=$((total_seconds + session_seconds))
                session_count=$((session_count + 1))
            fi
        done
    fi

    # 格式化輸出
    local hours=$((total_seconds / 3600))
    local minutes=$(((total_seconds % 3600) / 60))

    # 取得星期幾
    local weekday=$(get_chinese_weekday "$date")

    if [ $total_seconds -eq 0 ]; then
        echo "$date $weekday: 無記錄"
    else
        printf "%s %s: %dh %dm (%d sessions)\\n" "$date" "$weekday" "$hours" "$minutes" "$session_count"
    fi
}

# 計算日期範圍統計
calculate_range_total() {
    local start_date=$1
    local end_date=$2
    local grand_total=0
    local total_sessions=0

    echo "統計範圍: $start_date 至 $end_date"
    echo "----------------------------------------"

    # 遍歷日期範圍
    current_date="$start_date"
    while [ "$current_date" != "$end_date" ] || [ "$current_date" = "$end_date" ]; do
        # 檢查是否已超過結束日期
        if [[ "$current_date" > "$end_date" ]]; then
            break
        fi
        local total_seconds=0
        local session_count=0

        # 檢查活躍目錄（移除日期限制）
        for session_file in "$SESSIONS_DIR"/*.json; do
            if [ -f "$session_file" ]; then
                local session_date=$(jq -r '.date' "$session_file" 2>/dev/null)
                if [ "$session_date" = "$current_date" ]; then
                    local session_seconds=$(jq '.total_seconds // 0' "$session_file")
                    total_seconds=$((total_seconds + session_seconds))
                    session_count=$((session_count + 1))
                fi
            fi
        done

        # 檢查歸檔目錄
        if [ -d "$ARCHIVE_DIR/$current_date" ]; then
            for session_file in "$ARCHIVE_DIR/$current_date"/*.json; do
                if [ -f "$session_file" ]; then
                    local session_seconds=$(jq '.total_seconds // 0' "$session_file")
                    total_seconds=$((total_seconds + session_seconds))
                    session_count=$((session_count + 1))
                fi
            done
        fi

        # 只顯示有記錄的日期
        if [ $total_seconds -gt 0 ]; then
            local hours=$((total_seconds / 3600))
            local minutes=$(((total_seconds % 3600) / 60))
            local weekday=$(get_chinese_weekday "$current_date")
            printf "  %s %s: %2dh %2dm (%d sessions)\\n" "$current_date" "$weekday" "$hours" "$minutes" "$session_count"
            grand_total=$((grand_total + total_seconds))
            total_sessions=$((total_sessions + session_count))
        fi

        # 如果已經是結束日期，跳出循環
        if [ "$current_date" = "$end_date" ]; then
            break
        fi

        # 下一天
        if date --version >/dev/null 2>&1; then
            # GNU date
            current_date=$(date -d "$current_date + 1 day" +%Y-%m-%d)
        else
            # BSD date (macOS)
            current_date=$(date -j -v+1d -f "%Y-%m-%d" "$current_date" +%Y-%m-%d)
        fi
    done

    # 總計
    echo "----------------------------------------"
    local total_hours=$((grand_total / 3600))
    local total_minutes=$(((grand_total % 3600) / 60))
    printf "總計: %dh %dm (%d sessions)\\n" "$total_hours" "$total_minutes" "$total_sessions"
}

# 歸檔舊 session 檔案
archive_old_sessions() {
    local today=$(date +%Y-%m-%d)
    local moved_count=0
    local date_count=0

    echo "開始歸檔舊 session 檔案..."

    # 收集需要歸檔的日期
    local dates_to_archive=()
    for session_file in "$SESSIONS_DIR"/*.json; do
        if [ -f "$session_file" ]; then
            local session_date=$(jq -r '.date' "$session_file" 2>/dev/null)
            if [ -n "$session_date" ] && [ "$session_date" != "$today" ]; then
                # 檢查是否已在陣列中
                local found=0
                for existing_date in "${dates_to_archive[@]}"; do
                    if [ "$existing_date" = "$session_date" ]; then
                        found=1
                        break
                    fi
                done
                if [ $found -eq 0 ]; then
                    dates_to_archive+=("$session_date")
                fi
            fi
        fi
    done

    # 歸檔每個日期的檔案
    for date in "${dates_to_archive[@]}"; do
        local archive_path="$ARCHIVE_DIR/$date"
        mkdir -p "$archive_path"
        date_count=$((date_count + 1))

        # 移動該日期的所有檔案
        local files_moved=0
        for session_file in "$SESSIONS_DIR"/*.json; do
            if [ -f "$session_file" ]; then
                local session_date=$(jq -r '.date' "$session_file" 2>/dev/null)
                if [ "$session_date" = "$date" ]; then
                    mv "$session_file" "$archive_path/"
                    moved_count=$((moved_count + 1))
                    files_moved=$((files_moved + 1))
                fi
            fi
        done
        echo "  已歸檔 $date 的 $files_moved 個檔案"
    done

    if [ $moved_count -eq 0 ]; then
        echo "沒有需要歸檔的檔案"
    else
        echo "歸檔完成：移動了 $moved_count 個檔案到 $date_count 個日期目錄"
    fi
}

# 主程式
case "${1:-today}" in
    today)
        echo "=== 今日統計 ==="
        calculate_date_total "$(date +%Y-%m-%d)"
        ;;
    week)
        echo "=== 本週統計 ==="
        if date --version >/dev/null 2>&1; then
            # GNU date
            start_date=$(date -d "last monday" +%Y-%m-%d)
        else
            # BSD date (macOS)
            start_date=$(date -v-monday +%Y-%m-%d)
        fi
        end_date=$(date +%Y-%m-%d)
        calculate_range_total "$start_date" "$end_date"
        ;;
    month)
        echo "=== 本月統計 ==="
        start_date=$(date +%Y-%m-01)
        end_date=$(date +%Y-%m-%d)
        calculate_range_total "$start_date" "$end_date"
        ;;
    all)
        echo "=== 所有歷史統計 ==="
        # 找出最早的日期
        earliest_date=""

        # 檢查歸檔目錄
        if [ -d "$ARCHIVE_DIR" ]; then
            for dir in "$ARCHIVE_DIR"/*/; do
                if [ -d "$dir" ]; then
                    date_dir=$(basename "$dir")
                    if [ -z "$earliest_date" ] || [[ "$date_dir" < "$earliest_date" ]]; then
                        earliest_date="$date_dir"
                    fi
                fi
            done
        fi

        # 檢查當前 sessions
        for session_file in "$SESSIONS_DIR"/*.json; do
            if [ -f "$session_file" ]; then
                session_date=$(jq -r '.date' "$session_file" 2>/dev/null)
                if [ -n "$session_date" ]; then
                    if [ -z "$earliest_date" ] || [[ "$session_date" < "$earliest_date" ]]; then
                        earliest_date="$session_date"
                    fi
                fi
            fi
        done

        if [ -n "$earliest_date" ]; then
            calculate_range_total "$earliest_date" "$(date +%Y-%m-%d)"
        else
            echo "沒有找到任何記錄"
        fi
        ;;
    archive)
        echo "=== 歸檔舊 Session ==="
        archive_old_sessions
        ;;
    20[0-9][0-9]-[0-1][0-9]-[0-3][0-9])
        # 指定日期格式
        echo "=== 指定日期統計 ==="
        calculate_date_total "$1"
        ;;
    *)
        echo "用法: $0 [today|week|month|all|archive|YYYY-MM-DD]"
        exit 1
        ;;
esac

