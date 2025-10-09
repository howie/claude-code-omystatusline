# Claude Code 使用限制提醒功能

## 概述

在 statusline 上实时显示 Claude Code 的使用限制警告，帮助用户提前了解接近的限制，避免工作中断。

## 背景

Claude Code 有两个主要限制：
1. **Current Session**: 每个会话 4 小时后重置
2. **Weekly Limits**: 每周使用量限制（周四中午 12:00 重置）

## 目标

设计一个简洁但有效的提醒机制，在 statusline 上显示：
- Session 剩余时间警告
- Weekly 使用量警告
- 根据紧急程度使用不同的颜色和图标

## 设计方案

### 1. 数据源

Claude Code 通过 stdin 传递给 statusline script 的 JSON 包含以下字段：

```json
{
  "session_id": "当前会话 ID",
  "model": {
    "id": "模型 ID",
    "display_name": "显示名称",
    "info": {
      "session_cost": "会话成本",
      "daily_usage_percent": "每日使用百分比"
    }
  },
  "workspace": {
    "current_dir": "当前工作目录"
  },
  "transcript_path": "对话记录路径"
}
```

### 2. 警告级别系统

#### Session 时间警告

基于会话已运行时间和 4 小时限制：

| 剩余时间 | 级别 | 颜色 | 图标 | 显示 |
|---------|------|------|------|------|
| > 1 小时 | 正常 | - | - | 不显示 |
| 30-60 分钟 | 警告 | 黄色 | ⏰ | `⏰ 45m` |
| < 30 分钟 | 紧急 | 红色 | 🔴 | `🔴 15m` |

#### Weekly 使用量警告

基于 `daily_usage_percent` 字段（假设此字段代表周使用量）：

| 使用率 | 级别 | 颜色 | 图标 | 显示 |
|--------|------|------|------|------|
| < 70% | 正常 | - | - | 不显示 |
| 70-85% | 警告 | 黄色 | ⚠️ | `⚠️ 75%` |
| > 85% | 紧急 | 红色 | 🚨 | `🚨 92%` |

### 3. UI 设计

#### 显示位置

在现有 statusline 格式中插入警告信息：

```
[模型] 📂 项目 分支 | [限制警告] | Context使用 | 时间统计
```

#### 示例效果

**正常状态**（无警告）：
```
[💠 Sonnet 4.5] 📂 claude-code ⚡ main | ██████░░░░ 45% 90k | 2h30m
```

**单一警告**（Session 即将到期）：
```
[💠 Sonnet 4.5] 📂 claude-code ⚡ main | ⏰ 25m | ██████░░░░ 45% 90k | 3h35m
```

**双重警告**（Session + Weekly）：
```
[💠 Sonnet 4.5] 📂 claude-code ⚡ main | 🔴 10m ⚠️ 78% | ██████░░░░ 45% 90k | 3h50m
```

**紧急状态**：
```
[💠 Sonnet 4.5] 📂 claude-code ⚡ main | 🔴 5m 🚨 95% | ████████░░ 85% 170k | 3h55m
```

### 4. 实现细节

#### 4.1 更新 Input 结构体

在 `statusline.go` 中更新：

```go
type Input struct {
    Model struct {
        DisplayName string `json:"display_name"`
        Info struct {
            SessionCost       float64 `json:"session_cost"`
            DailyUsagePercent float64 `json:"daily_usage_percent"`
        } `json:"info"`
    } `json:"model"`
    SessionID      string `json:"session_id"`
    Workspace      struct {
        CurrentDir string `json:"current_dir"`
    } `json:"workspace"`
    TranscriptPath string `json:"transcript_path,omitempty"`
}
```

#### 4.2 新增函数

**计算 Session 剩余时间**

```go
func calculateSessionRemaining(sessionID string) (minutes int, level string) {
    // 从 session-tracker 读取会话开始时间
    // 计算已运行时间
    // 4小时限制 - 已运行时间 = 剩余时间
    // 返回剩余分钟数和警告级别
}
```

**检查使用限制**

```go
func checkUsageLimits(input Input, sessionID string) string {
    var warnings []string

    // 检查 Session 时间
    remaining, level := calculateSessionRemaining(sessionID)
    if level == "warning" {
        warnings = append(warnings, formatTimeWarning(remaining, "warning"))
    } else if level == "critical" {
        warnings = append(warnings, formatTimeWarning(remaining, "critical"))
    }

    // 检查 Weekly 使用量
    usagePercent := input.Model.Info.DailyUsagePercent
    if usagePercent >= 85 {
        warnings = append(warnings, formatUsageWarning(usagePercent, "critical"))
    } else if usagePercent >= 70 {
        warnings = append(warnings, formatUsageWarning(usagePercent, "warning"))
    }

    if len(warnings) > 0 {
        return " | " + strings.Join(warnings, " ")
    }
    return ""
}
```

**格式化警告**

```go
func formatTimeWarning(minutes int, level string) string {
    var icon, color string
    if level == "critical" {
        icon = "🔴"
        color = ColorCtxRed
    } else {
        icon = "⏰"
        color = ColorCtxGold
    }
    return fmt.Sprintf("%s%s %dm%s", color, icon, minutes, ColorReset)
}

func formatUsageWarning(percent float64, level string) string {
    var icon, color string
    if level == "critical" {
        icon = "🚨"
        color = ColorCtxRed
    } else {
        icon = "⚠️"
        color = ColorCtxGold
    }
    return fmt.Sprintf("%s%s %.0f%%%s", color, icon, percent, ColorReset)
}
```

#### 4.3 整合到主输出

修改 `main()` 函数中的输出部分：

```go
// 检查使用限制
limitWarnings := checkUsageLimits(input, input.SessionID)

// 输出状态列
fmt.Printf("%s[%s] 📂 %s%s%s%s | %s%s\n",
    ColorReset, modelDisplay, projectName, gitBranch,
    limitWarnings,  // 新增：限制警告
    contextUsage, totalHours, ColorReset)
```

### 5. 颜色定义

需要的颜色常量（已在现有代码中）：

```go
const (
    ColorCtxGreen = "\033[38;2;108;167;108m"  // 绿色（正常）
    ColorCtxGold  = "\033[38;2;188;155;83m"   // 黄色（警告）
    ColorCtxRed   = "\033[38;2;185;102;82m"   // 红色（紧急）
)
```

### 6. 测试计划

#### 6.1 Session 时间测试

1. 创建新会话，验证无警告
2. 修改 session 开始时间到 3 小时前，验证黄色警告
3. 修改到 3.5 小时前，验证红色警告

#### 6.2 Usage 测试

使用测试 JSON 输入：

```json
{
  "model": {
    "display_name": "Sonnet 4.5",
    "info": {
      "daily_usage_percent": 65
    }
  },
  "session_id": "test-session",
  "workspace": {
    "current_dir": "/test"
  }
}
```

测试不同的 `daily_usage_percent` 值：65, 75, 90

### 7. 边界情况处理

1. **缺失数据字段**
   - 如果 `model.info` 不存在，跳过 usage 警告
   - 如果无法读取 session 数据，跳过时间警告

2. **数据异常**
   - Usage percent > 100%：视为 100%
   - Session 时间为负：不显示警告

3. **性能考虑**
   - Session 时间计算可以加入缓存（5秒）
   - 避免每次都重新计算

### 8. 未来改进

1. **更精确的限制信息**
   - 如果 API 提供更详细的限制数据，可以显示具体的 token 剩余量
   - 区分不同模型的限制

2. **自定义阈值**
   - 允许用户配置警告阈值
   - 通过配置文件设置

3. **历史趋势**
   - 显示使用趋势（上升/下降）
   - 预测何时会达到限制

## 文件修改清单

- `~/.claude/statusline.go`
  - 更新 `Input` 结构体
  - 新增 `calculateSessionRemaining()` 函数
  - 新增 `checkUsageLimits()` 函数
  - 新增 `formatTimeWarning()` 函数
  - 新增 `formatUsageWarning()` 函数
  - 修改 `main()` 函数输出部分

## 依赖

- 现有的 session-tracker 系统
- Claude Code 的 JSON 输入格式（需要验证 `model.info` 字段是否存在）

## 注意事项

1. **验证数据格式**：需要先验证 Claude Code 实际传递的 JSON 格式，确认 `model.info.daily_usage_percent` 字段确实存在

2. **Session 重置时机**：需要确认 4 小时限制的具体计算方式（是否包含空闲时间）

3. **跨平台兼容性**：确保 emoji 图标在不同终端上正确显示
