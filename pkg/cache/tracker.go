package cache

import (
	"fmt"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

// CacheInfo 代表快取命中率資訊
type CacheInfo struct {
	HitRate    int // 0-100 百分比
	CacheRead  int // cache_read_input_tokens
	TotalInput int // 三種 token 的總和
}

// Calculate 從 transcript 行計算快取命中率
func Calculate(lines []transcript.Line) *CacheInfo {
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		if l.Parsed == nil {
			continue
		}

		if isSide, ok := l.Parsed["isSidechain"].(bool); ok && isSide {
			continue
		}

		msg, ok := l.Parsed["message"].(map[string]interface{})
		if !ok {
			continue
		}

		usage, ok := msg["usage"].(map[string]interface{})
		if !ok {
			continue
		}

		var inputTokens, cacheRead, cacheCreation float64

		if v, ok := usage["input_tokens"].(float64); ok {
			inputTokens = v
		}
		if v, ok := usage["cache_read_input_tokens"].(float64); ok {
			cacheRead = v
		}
		if v, ok := usage["cache_creation_input_tokens"].(float64); ok {
			cacheCreation = v
		}

		total := inputTokens + cacheRead + cacheCreation
		if total <= 0 {
			continue
		}

		hitRate := int(cacheRead * 100.0 / total)
		return &CacheInfo{
			HitRate:    hitRate,
			CacheRead:  int(cacheRead),
			TotalInput: int(total),
		}
	}

	return nil
}

// Format 格式化快取命中率顯示
func Format(info *CacheInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("Cache %d%%", info.HitRate)
}
