package git

import (
	"os"
	"testing"
)

// TestMain 清除繼承自外部環境的 GIT_* 變數，確保測試隔離。
//
// 當測試在 git hook（例如 pre-commit）或某些 CI 環境下執行時，git 會 export
// GIT_DIR、GIT_WORK_TREE、GIT_INDEX_FILE 等變數指向「正在 commit 的那個 repo」。
// 這些變數的優先級高於 `git -C <dir>`，導致 GetBranch 對臨時測試 repo 的查詢
// 被重導到真實 repo，回傳錯誤的分支名（測試會看到當前工作分支而非 test-branch）。
//
// 在所有測試開始前清掉這些變數，可讓 `git -C <tmpDir>` 真正作用於 tmpDir。
func TestMain(m *testing.M) {
	for _, key := range []string{
		"GIT_DIR",
		"GIT_WORK_TREE",
		"GIT_INDEX_FILE",
		"GIT_COMMON_DIR",
		"GIT_OBJECT_DIRECTORY",
		"GIT_PREFIX",
	} {
		_ = os.Unsetenv(key)
	}
	os.Exit(m.Run())
}
