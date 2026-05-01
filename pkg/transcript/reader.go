package transcript

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
)

// Line 代表 transcript 中的一行
type Line struct {
	Raw    string
	Parsed map[string]interface{}
}

// ReadTail 讀取 transcript 檔案的最後 n 行並解析 JSON
func ReadTail(path string, n int) ([]Line, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	// 使用 1MB buffer 處理長 JSON 行
	const maxScanTokenSize = 1024 * 1024
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	start := len(allLines) - n
	if start < 0 {
		start = 0
	}
	rawLines := allLines[start:]

	lines := make([]Line, 0, len(rawLines))
	for _, raw := range rawLines {
		l := Line{Raw: raw}
		if raw != "" {
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
				l.Parsed = parsed
			}
		}
		lines = append(lines, l)
	}

	return lines, nil
}

// ReadAll 讀取 transcript 全部行並解析（用於需要完整記錄的情境）
func ReadAll(path string) ([]Line, error) {
	return ReadTail(path, math.MaxInt)
}

// FilterBySession 過濾出特定 session 的非 sidechain 行
func FilterBySession(lines []Line, sessionID string) []Line {
	var result []Line
	for _, l := range lines {
		if l.Parsed == nil {
			continue
		}
		if isSide, ok := l.Parsed["isSidechain"].(bool); ok && isSide {
			continue
		}
		if sid, ok := l.Parsed["sessionId"].(string); ok && sid == sessionID {
			result = append(result, l)
		}
	}
	return result
}
