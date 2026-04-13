package cache

import (
	"testing"

	"github.com/howie/claude-code-omystatusline/pkg/transcript"
)

func makeLine(raw string, parsed map[string]interface{}) transcript.Line {
	return transcript.Line{Raw: raw, Parsed: parsed}
}

func TestCalculateNormal(t *testing.T) {
	lines := []transcript.Line{
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":                float64(20),
					"cache_read_input_tokens":     float64(80),
					"cache_creation_input_tokens": float64(0),
				},
			},
			"isSidechain": false,
		}),
	}

	info := Calculate(lines)
	if info == nil {
		t.Fatal("expected non-nil CacheInfo")
	}
	if info.HitRate != 80 {
		t.Errorf("expected HitRate=80, got %d", info.HitRate)
	}
	if info.CacheRead != 80 {
		t.Errorf("expected CacheRead=80, got %d", info.CacheRead)
	}
	if info.TotalInput != 100 {
		t.Errorf("expected TotalInput=100, got %d", info.TotalInput)
	}
}

func TestCalculateSkipsSidechain(t *testing.T) {
	lines := []transcript.Line{
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":                float64(50),
					"cache_read_input_tokens":     float64(50),
					"cache_creation_input_tokens": float64(0),
				},
			},
			"isSidechain": false,
		}),
		// 最後一行是 sidechain，應被跳過
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":                float64(100),
					"cache_read_input_tokens":     float64(0),
					"cache_creation_input_tokens": float64(0),
				},
			},
			"isSidechain": true,
		}),
	}

	info := Calculate(lines)
	if info == nil {
		t.Fatal("expected non-nil CacheInfo")
	}
	// 應該取得第一行（非 sidechain）的資料
	if info.HitRate != 50 {
		t.Errorf("expected HitRate=50, got %d", info.HitRate)
	}
}

func TestCalculateNoUsage(t *testing.T) {
	lines := []transcript.Line{
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"role": "user",
			},
		}),
		makeLine("", nil),
	}

	info := Calculate(lines)
	if info != nil {
		t.Errorf("expected nil CacheInfo, got %+v", info)
	}
}

func TestCalculateZeroTotal(t *testing.T) {
	lines := []transcript.Line{
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":                float64(0),
					"cache_read_input_tokens":     float64(0),
					"cache_creation_input_tokens": float64(0),
				},
			},
			"isSidechain": false,
		}),
	}

	info := Calculate(lines)
	if info != nil {
		t.Errorf("expected nil CacheInfo for zero total, got %+v", info)
	}
}

func TestCalculateAllCacheRead(t *testing.T) {
	lines := []transcript.Line{
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":                float64(0),
					"cache_read_input_tokens":     float64(100),
					"cache_creation_input_tokens": float64(0),
				},
			},
			"isSidechain": false,
		}),
	}

	info := Calculate(lines)
	if info == nil {
		t.Fatal("expected non-nil CacheInfo")
	}
	if info.HitRate != 100 {
		t.Errorf("expected HitRate=100, got %d", info.HitRate)
	}
}

func TestCalculateNoCacheRead(t *testing.T) {
	lines := []transcript.Line{
		makeLine("", map[string]interface{}{
			"message": map[string]interface{}{
				"usage": map[string]interface{}{
					"input_tokens":                float64(100),
					"cache_read_input_tokens":     float64(0),
					"cache_creation_input_tokens": float64(50),
				},
			},
			"isSidechain": false,
		}),
	}

	info := Calculate(lines)
	if info == nil {
		t.Fatal("expected non-nil CacheInfo")
	}
	if info.HitRate != 0 {
		t.Errorf("expected HitRate=0, got %d", info.HitRate)
	}
}

func TestCalculateEmptyLines(t *testing.T) {
	info := Calculate(nil)
	if info != nil {
		t.Errorf("expected nil for nil lines, got %+v", info)
	}

	info = Calculate([]transcript.Line{})
	if info != nil {
		t.Errorf("expected nil for empty lines, got %+v", info)
	}
}

func TestFormat(t *testing.T) {
	if got := Format(nil); got != "" {
		t.Errorf("Format(nil) = %q, want empty", got)
	}

	info := &CacheInfo{HitRate: 85}
	if got := Format(info); got != "Cache 85%" {
		t.Errorf("Format({HitRate:85}) = %q, want %q", got, "Cache 85%")
	}

	info = &CacheInfo{HitRate: 0}
	if got := Format(info); got != "Cache 0%" {
		t.Errorf("Format({HitRate:0}) = %q, want %q", got, "Cache 0%")
	}

	info = &CacheInfo{HitRate: 100}
	if got := Format(info); got != "Cache 100%" {
		t.Errorf("Format({HitRate:100}) = %q, want %q", got, "Cache 100%")
	}
}
