package apilimits

import (
	"strings"
	"testing"
	"time"
)

// 說明：formatRemaining 用 time.Until(t)，實際剩餘時間會比設定的 offset 少數微秒，
// int() 截斷會把「正好 2h」變成「1h59m」。因此測試一律加 30s buffer，
// 讓截斷後的整數落在預期值（30s 遠小於分鐘解析度，但足以蓋過微秒誤差）。
const tinyBuffer = 30 * time.Second

func TestFromRateLimits(t *testing.T) {
	now := time.Now()
	fiveHourReset := now.Add(2*time.Hour + tinyBuffer).Unix()
	sevenDayReset := now.Add(48*time.Hour + tinyBuffer).Unix()

	info := FromRateLimits(
		RateLimitWindow{UsedPercentage: 23.5, ResetsAtUnix: fiveHourReset},
		RateLimitWindow{UsedPercentage: 41.2, ResetsAtUnix: sevenDayReset},
	)
	if info == nil {
		t.Fatal("FromRateLimits returned nil")
	}

	// used_percentage 已是 0-100，直接取整數（不像 OAuth 的 0-1 需 ×100）。
	if info.FiveHourPct != 23 {
		t.Errorf("FiveHourPct = %d, want 23", info.FiveHourPct)
	}
	if info.SevenDayPct != 41 {
		t.Errorf("SevenDayPct = %d, want 41", info.SevenDayPct)
	}
	if info.FiveHourReset != "2h" {
		t.Errorf("FiveHourReset = %q, want \"2h\"", info.FiveHourReset)
	}
	if info.SevenDayReset != "2d" {
		t.Errorf("SevenDayReset = %q, want \"2d\"", info.SevenDayReset)
	}
	if info.LimitReached {
		t.Error("LimitReached should be false for sub-100% usage")
	}
}

func TestFromRateLimitsLimitReached(t *testing.T) {
	reset := time.Now().Add(time.Hour + tinyBuffer).Unix()
	info := FromRateLimits(
		RateLimitWindow{UsedPercentage: 100, ResetsAtUnix: reset},
		RateLimitWindow{UsedPercentage: 50, ResetsAtUnix: reset},
	)
	if !info.LimitReached {
		t.Error("LimitReached should be true when a window hits 100%")
	}
}

func TestFromRateLimitsZeroUsage(t *testing.T) {
	// 0% 使用但有 resets_at：合法狀態（session 剛開始），不應視為缺資料。
	reset := time.Now().Add(time.Hour + tinyBuffer).Unix()
	info := FromRateLimits(
		RateLimitWindow{UsedPercentage: 0, ResetsAtUnix: reset},
		RateLimitWindow{UsedPercentage: 0, ResetsAtUnix: reset},
	)
	if info.FiveHourPct != 0 || info.SevenDayPct != 0 {
		t.Errorf("expected 0%% usage, got 5h=%d 7d=%d", info.FiveHourPct, info.SevenDayPct)
	}
	if info.FiveHourReset == "" {
		t.Error("FiveHourReset should be populated when resets_at > 0")
	}
}

func TestFormatResetUnix(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		epoch int64
		want  string
	}{
		{"zero epoch yields empty", 0, ""},
		{"negative epoch yields empty", -1, ""},
		{"past time yields now", now.Add(-time.Hour).Unix(), "now"},
		{"45 minutes ahead", now.Add(45*time.Minute + tinyBuffer).Unix(), "45m"},
		{"2h30m ahead", now.Add(2*time.Hour + 30*time.Minute + tinyBuffer).Unix(), "2h30m"},
		{"3 days ahead", now.Add(72*time.Hour + tinyBuffer).Unix(), "3d"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatResetUnix(tc.epoch)
			if got != tc.want {
				t.Errorf("formatResetUnix(%d) = %q, want %q", tc.epoch, got, tc.want)
			}
		})
	}
}

func TestFormatFromRateLimits(t *testing.T) {
	reset := time.Now().Add(3*time.Hour + tinyBuffer).Unix()
	info := FromRateLimits(
		RateLimitWindow{UsedPercentage: 25, ResetsAtUnix: reset},
		RateLimitWindow{UsedPercentage: 40, ResetsAtUnix: reset},
	)
	out := Format(info)

	for _, want := range []string{"5h: 25%", "7d: 40%"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format output %q missing %q", out, want)
		}
	}
}

func TestFormatNil(t *testing.T) {
	if got := Format(nil); got != "" {
		t.Errorf("Format(nil) = %q, want empty", got)
	}
}
