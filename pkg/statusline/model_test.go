package statusline

import (
	"encoding/json"
	"testing"
)

// TestInputWorktreeParsing 驗證 worktree 欄位能正確從官方 Claude Code statusline
// schema 解析。此測試對應 #29：先前 OriginalCwd 誤用 "original_repo_dir" key，
// 官方從未提供該 key，導致該欄位永遠為空字串。
func TestInputWorktreeParsing(t *testing.T) {
	// 取自官方 statusline JSON schema（worktree 區段）。
	raw := `{
		"worktree": {
			"name": "my-feature",
			"path": "/path/to/.claude/worktrees/my-feature",
			"branch": "worktree-my-feature",
			"original_cwd": "/path/to/project",
			"original_branch": "main"
		}
	}`

	var input Input
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Name", input.Worktree.Name, "my-feature"},
		{"Path", input.Worktree.Path, "/path/to/.claude/worktrees/my-feature"},
		{"Branch", input.Worktree.Branch, "worktree-my-feature"},
		{"OriginalCwd", input.Worktree.OriginalCwd, "/path/to/project"},
		{"OriginalBranch", input.Worktree.OriginalBranch, "main"},
	}
	for _, tc := range tests {
		if tc.got != tc.want {
			t.Errorf("Worktree.%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

// TestInputRateLimitsParsing 驗證 rate_limits 欄位能正確從官方 Claude Code statusline
// schema 解析。此測試對應 #30：used_percentage 為 0-100 浮點，resets_at 為 Unix epoch。
func TestInputRateLimitsParsing(t *testing.T) {
	// 取自官方 statusline JSON schema（rate_limits 區段）。
	raw := `{
		"rate_limits": {
			"five_hour": {
				"used_percentage": 23.5,
				"resets_at": 1738425600
			},
			"seven_day": {
				"used_percentage": 41.2,
				"resets_at": 1738857600
			}
		}
	}`

	var input Input
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got := input.RateLimits.FiveHour.UsedPercentage; got != 23.5 {
		t.Errorf("FiveHour.UsedPercentage = %v, want 23.5", got)
	}
	if got := input.RateLimits.FiveHour.ResetsAt; got != 1738425600 {
		t.Errorf("FiveHour.ResetsAt = %d, want 1738425600", got)
	}
	if got := input.RateLimits.SevenDay.UsedPercentage; got != 41.2 {
		t.Errorf("SevenDay.UsedPercentage = %v, want 41.2", got)
	}
	if got := input.RateLimits.SevenDay.ResetsAt; got != 1738857600 {
		t.Errorf("SevenDay.ResetsAt = %d, want 1738857600", got)
	}
}
