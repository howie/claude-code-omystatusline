package terminal

import (
	"testing"
)

func TestWidthFromEnv(t *testing.T) {
	t.Setenv("COLUMNS", "80")
	// ioctl 在非 TTY 環境會失敗，應 fallback 到 COLUMNS
	w := Width()
	if w <= 0 {
		t.Errorf("Width() = %d, want > 0", w)
	}
}

func TestWidthDefault(t *testing.T) {
	// 清除 COLUMNS，強迫走預設值路徑
	t.Setenv("COLUMNS", "")

	w := Width()
	if w <= 0 {
		t.Errorf("Width() default = %d, want > 0", w)
	}
}

func TestWidthInvalidColumns(t *testing.T) {
	t.Setenv("COLUMNS", "not-a-number")
	w := Width()
	if w <= 0 {
		t.Errorf("Width() with invalid COLUMNS = %d, want > 0", w)
	}
}
