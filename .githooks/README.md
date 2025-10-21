# Git Hooks for Claude Code omystatusline

This directory contains Git hooks to help maintain code quality.

## Available Hooks

### pre-push

Runs before `git push` to ensure code quality. Performs the following checks:

1. **Working Directory Status**: Checks for uncommitted changes
2. **Go Compilation**: Ensures Go code compiles without errors
3. **Go Tests**: Runs all Go tests
4. **Shell Script Syntax**: Validates shell script syntax
5. **Install Script Test**: Tests the installer script

## Installation

### Quick Install

Run from the project root:

```bash
make install-hooks
```

Or manually:

```bash
./.githooks/install-hooks.sh
```

### Manual Installation

Copy the hooks to `.git/hooks/`:

```bash
cp .githooks/pre-push .git/hooks/pre-push
chmod +x .git/hooks/pre-push
```

## Usage

Once installed, the hooks run automatically:

```bash
git push origin branch-name
```

The pre-push hook will run checks before the push. If any check fails, you'll see:

```
╔════════════════════════════════════════════════════════════════╗
║  Pre-Push Checks - Claude Code omystatusline              ║
╚════════════════════════════════════════════════════════════════╝

ℹ 檢查 1/5: 工作目錄狀態
✓ 工作目錄乾淨

ℹ 檢查 2/5: Go 代碼編譯
✓ Go 代碼編譯成功

ℹ 檢查 3/5: Go 測試執行
✓ Go 測試通過

ℹ 檢查 4/5: Shell 腳本語法檢查
✓ install.sh 語法正確
✓ statusline-wrapper.sh 語法正確
✓ statusline.sh 語法正確

ℹ 檢查 5/5: 安裝腳本乾跑測試
✓ 安裝腳本載入測試通過

════════════════════════════════════════════════════════════════
✓ 所有檢查通過！
════════════════════════════════════════════════════════════════
```

## Bypassing Hooks

If you need to bypass the hooks (not recommended):

```bash
git push --no-verify origin branch-name
```

## Uninstallation

Remove the hooks:

```bash
make uninstall-hooks
```

Or manually:

```bash
rm .git/hooks/pre-push
```

## Customization

You can customize the checks by editing `.githooks/pre-push`. The hook is a bash script that exits with:
- `0` - All checks passed, allow push
- `1` - Checks failed, block push

## Development

### Testing the Hook

Test the hook without pushing:

```bash
.git/hooks/pre-push origin https://github.com/user/repo
```

### Adding New Checks

Add new checks to `.githooks/pre-push`:

1. Add a new check section
2. Update the check counter
3. Update the header to reflect the new total
4. Increment `FAILED_CHECKS` if the check fails

Example:

```bash
# ============================================================================
# 檢查 6/6: 新檢查
# ============================================================================
print_info "檢查 6/6: 新檢查描述"

if your_check_command; then
    print_success "新檢查通過"
else
    print_error "新檢查失敗"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi
```

## Troubleshooting

### Hook not running

Make sure the hook is executable:

```bash
chmod +x .git/hooks/pre-push
```

### Go not found

If you don't have Go installed, the compilation and test checks will be skipped with a warning.

### Permission denied

The hook script needs execute permission:

```bash
chmod +x .githooks/pre-push
```

## Best Practices

1. **Always run hooks**: Don't bypass hooks unless absolutely necessary
2. **Fix failures**: Address check failures before pushing
3. **Keep hooks updated**: Pull the latest hooks after fetching changes
4. **Test locally**: Run checks manually before committing

## Related Commands

```bash
make test              # Run tests manually
make build             # Build the project
make install-hooks     # Install Git hooks
make uninstall-hooks   # Remove Git hooks
```
