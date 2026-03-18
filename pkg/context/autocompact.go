package context

import (
	"fmt"
	"strings"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// AutocompactInfo 代表自動壓縮偵測結果
type AutocompactInfo struct {
	Detected bool // 是否偵測到 context 壓縮
	Count    int  // 壓縮次數
}

// DetectAutocompact 從 transcript 行中偵測 context 壓縮事件
func DetectAutocompact(lines []transcript.Line) *AutocompactInfo {
	info := &AutocompactInfo{}

	for _, l := range lines {
		if l.Parsed == nil {
			continue
		}

		// 檢查 type == "summary" 表示壓縮事件
		if msgType, ok := l.Parsed["type"].(string); ok && msgType == "summary" {
			info.Detected = true
			info.Count++
			continue
		}

		// 也檢查 message 中的 system 角色包含壓縮相關內容
		if msg, ok := l.Parsed["message"].(map[string]interface{}); ok {
			role, _ := msg["role"].(string)
			if role == "system" {
				if content, ok := msg["content"].(string); ok {
					if strings.Contains(content, "autocompact") ||
						strings.Contains(content, "context window") ||
						strings.Contains(content, "compressed") {
						info.Detected = true
						info.Count++
					}
				}
			}
		}
	}

	return info
}

// FormatAutocompact 格式化壓縮偵測結果
func FormatAutocompact(info *AutocompactInfo) string {
	if info == nil || !info.Detected {
		return ""
	}
	if info.Count > 1 {
		return fmt.Sprintf("⚠ compressed ×%d", info.Count)
	}
	return "⚠ compressed"
}
