package statusline

import (
	"sort"
	"strings"
	"unicode"
)

// Segment 代表 status line 的一個顯示段落。
// Priority 越小越重要，截斷時從大到小移除。
type Segment struct {
	Content  string
	Priority int
}

// VisibleWidth 計算字串的可見欄位寬度，
// 跳過 ANSI escape sequences，emoji/寬字元算 2 欄。
func VisibleWidth(s string) int {
	width := 0
	inEscape := false
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inEscape {
			// CSI 序列以字母結束（@-~，即 0x40-0x7E）
			if r >= 0x40 && r <= 0x7E {
				inEscape = false
			}
			continue
		}
		if r == '\033' && i+1 < len(runes) && runes[i+1] == '[' {
			inEscape = true
			i++ // 跳過 '['
			continue
		}
		width += runeWidth(r)
	}
	return width
}

// runeWidth 回傳單一 rune 的顯示寬度。
func runeWidth(r rune) int {
	// 控制字元
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 0
	}
	// 組合用字元（零寬）
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) {
		return 0
	}
	// 全形與寬字元範圍（含 CJK、emoji）
	if isWide(r) {
		return 2
	}
	return 1
}

// isWide 判斷是否為佔 2 欄的字元。
func isWide(r rune) bool {
	// East Asian Wide / Fullwidth 範圍
	switch {
	case r >= 0x1100 && r <= 0x115F: // Hangul Jamo
		return true
	case r >= 0x2E80 && r <= 0x303E: // CJK Radicals / Kangxi
		return true
	case r >= 0x3041 && r <= 0x33BF: // Japanese / CJK
		return true
	case r >= 0x33FF && r <= 0xA4CF: // CJK Compat tail + Extension A + Unified Ideographs + Yi
		return true
	case r >= 0xA960 && r <= 0xA97F: // Hangul
		return true
	case r >= 0xAC00 && r <= 0xD7FF: // Hangul Syllables
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility
		return true
	case r >= 0xFE10 && r <= 0xFE1F: // Vertical Forms
		return true
	case r >= 0xFE30 && r <= 0xFE6F: // CJK Compatibility Forms
		return true
	case r >= 0xFF00 && r <= 0xFF60: // Fullwidth Forms
		return true
	case r >= 0xFFE0 && r <= 0xFFE6: // Fullwidth Signs
		return true
	case r >= 0x1B000 && r <= 0x1B0FF: // Kana Supplement
		return true
	case r >= 0x1F004 && r <= 0x1F0CF: // Playing Cards etc.
		return true
	case r >= 0x1F300 && r <= 0x1F9FF: // Misc Symbols & Emoji (📂 💛 💰 ⚡ 🌸 💠 🔇 💬 ⚠ etc.)
		return true
	case r >= 0x20000 && r <= 0x2FFFD: // CJK Extension B+
		return true
	case r >= 0x30000 && r <= 0x3FFFD: // CJK Extension G+
		return true
	}
	return false
}

// TruncateLine 將 segments 依優先級截斷至 maxWidth 欄寬以內。
// 若全部段落都放得下，直接串接回傳。
// 若超出，從優先級最高的數字（最不重要）開始移除段落，
// 直到符合寬度，或下一個待移除段落的優先級 ≤ 1 為止。
// Priority 0（不可見控制字元，如 ColorReset）與 Priority 1 同樣受保護。
// 被截斷時在末尾加 「…」。
func TruncateLine(segments []Segment, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 120
	}

	// 過濾空段落
	var active []Segment
	for _, s := range segments {
		if s.Content != "" {
			active = append(active, s)
		}
	}

	// 計算總寬度
	total := 0
	for _, s := range active {
		total += VisibleWidth(s.Content)
	}

	if total <= maxWidth {
		return joinSegments(active)
	}

	// 按優先級降序排列（數字大 = 優先級低 = 先移除）
	type indexed struct {
		seg Segment
		idx int
	}
	indexed_segs := make([]indexed, len(active))
	for i, s := range active {
		indexed_segs[i] = indexed{s, i}
	}
	sort.SliceStable(indexed_segs, func(i, j int) bool {
		return indexed_segs[i].seg.Priority > indexed_segs[j].seg.Priority
	})

	ellipsis := "…"
	ellipsisWidth := VisibleWidth(ellipsis)

	// 逐步移除最低優先級段落
	removed := make(map[int]bool)
	for _, item := range indexed_segs {
		if total+ellipsisWidth <= maxWidth {
			break
		}
		// 不移除最高優先級（Priority == 1）
		if item.seg.Priority <= 1 {
			break
		}
		removed[item.idx] = true
		total -= VisibleWidth(item.seg.Content)
	}

	// 重組保留的段落（維持原順序）
	var kept []Segment
	for i, s := range active {
		if !removed[i] {
			kept = append(kept, s)
		}
	}

	result := joinSegments(kept)
	if len(removed) > 0 {
		result += ellipsis
	}
	return result
}

func joinSegments(segs []Segment) string {
	var sb strings.Builder
	for _, s := range segs {
		sb.WriteString(s.Content)
	}
	return sb.String()
}

// WrapLine 將 segments 依終端寬度智慧換行。
// 超出 maxWidth 時，在段落邊界換行到第二行（前綴 1 格縮排）；
// 若兩行仍不足，對第二行以優先級截斷。
// 若全部放得下，直接回傳單行（與 TruncateLine 相同）。
// 特殊情形：若首個段落本身即超出 maxWidth，仍強制放入第一行不截斷（最小保證）。
func WrapLine(segments []Segment, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 120
	}

	// 過濾空段落
	var active []Segment
	for _, s := range segments {
		if s.Content != "" {
			active = append(active, s)
		}
	}
	if len(active) == 0 {
		return ""
	}

	// 計算總寬度
	total := 0
	for _, s := range active {
		total += VisibleWidth(s.Content)
	}

	// 全部放得下：直接回傳單行
	if total <= maxWidth {
		return joinSegments(active)
	}

	// 分行：按順序將段落填入第一行，超出後移至第二行
	var line1, line2 []Segment
	currentWidth := 0
	for _, s := range active {
		w := VisibleWidth(s.Content)
		if len(line2) == 0 && currentWidth+w <= maxWidth {
			line1 = append(line1, s)
			currentWidth += w
		} else {
			line2 = append(line2, s)
		}
	}

	// 若第一行完全為空（首個段落本身超出），強制放入
	if len(line1) == 0 && len(line2) > 0 {
		line1 = append(line1, line2[0])
		line2 = line2[1:]
	}

	// 若無第二行，回傳單行
	if len(line2) == 0 {
		return joinSegments(line1)
	}

	// 第二行首個段落去掉前導分隔符（避免孤立 | ）
	first := line2[0]
	line2[0] = Segment{Content: stripLeadingDivider(first.Content), Priority: first.Priority}

	const line2Prefix = "↳ "
	prefixWidth := VisibleWidth(line2Prefix)

	// 計算第二行寬度
	line2Total := prefixWidth
	for _, s := range line2 {
		line2Total += VisibleWidth(s.Content)
	}

	// 若第二行超出：以優先級截斷
	var line2Str string
	if line2Total > maxWidth {
		line2MaxWidth := maxWidth - prefixWidth
		if line2MaxWidth <= 0 {
			line2MaxWidth = maxWidth
		}
		line2Str = line2Prefix + TruncateLine(line2, line2MaxWidth)
	} else {
		line2Str = line2Prefix + joinSegments(line2)
	}

	return joinSegments(line1) + "\n" + line2Str
}

// stripLeadingDivider 去掉段落開頭的常見分隔符前綴（含兩側空白）。
func stripLeadingDivider(s string) string {
	for _, prefix := range []string{" | ", " \ue0b1 ", " \ue0b0 ", " │ "} {
		if strings.HasPrefix(s, prefix) {
			return s[len(prefix):]
		}
	}
	return s
}
