# NOFX Bug 修复文档

> 本文档记录 NOFX 项目中发现的所有 Bug 及其修复方案，方便开发和维护人员快速查找和解决问题。

---

## 目录

1. [Bug 分类说明](#bug-分类说明)
2. [Bug 记录](#bug-记录)
3. [修复流程](#修复流程)

---

## Bug 分类说明

### 严重程度

| 级别 | 说明 | 标记 |
|------|------|------|
| **P0** | 严重阻塞，系统无法使用 | 🔴 |
| **P1** | 重要功能异常，影响核心流程 | 🟠 |
| **P2** | 一般功能异常，有替代方案 | 🟡 |
| **P3** | 轻微问题，不影响使用 | 🟢 |

### 模块分类

- **后端 (Backend)**: Go API 服务
- **前端 (Frontend)**: React 界面
- **AI 模型 (AI Model)**: AI 模型调用
- **交易所 (Exchange)**: 交易所连接
- **数据库 (Database)**: SQLite 数据库
- **部署 (Deploy)**: Docker 部署相关

---

## Bug 记录

### 2025-12-26

---

#### BUG-001: API Gateway 选项不显示

**模块**: 后端 / 前端
**严重程度**: P1
**状态**: ✅ 已修复

**问题描述**:
在添加 AI 模型时，下拉菜单中没有显示 "API Gateway (中转服务)" 选项。

**原因分析**:
1. 后端 `api/server.go` 中的 `handleGetSupportedModels` 函数未包含 gateway 配置
2. Docker 构建缓存导致新代码未生效
3. 前端 TypeScript 类型定义缺少 gateway 相关字段

**解决方案**:
1. 在 `api/server.go` 中添加 gateway provider 配置:
```go
{
    "id": "gateway",
    "name": "API Gateway (中转服务)",
    "provider": "gateway",
    "defaultModel": "gpt-4o-mini",
    "description": "Unified API gateway like zenmux.ai, openrouter.ai, etc.",
    "presets": []map[string]string{
        {"id": "zenmux", "name": "ZenMux.ai", "url": "https://api.zenmux.ai/v1"},
        {"id": "openrouter", "name": "OpenRouter.ai", "url": "https://openrouter.ai/api/v1"},
        {"id": "custom", "name": "Custom Gateway", "url": ""},
    },
}
```

2. 更新前端 `types.ts` 添加 gateway 相关字段:
```typescript
export interface AIModel {
  // ... 现有字段
  description?: string
  presets?: GatewayPreset[]
  selectedPreset?: string
}
```

3. 更新 `ModelConfigModal` 组件支持 gateway 配置界面

4. 创建 gateway 图标文件 `web/public/icons/gateway.svg`

5. 使用 `--no-cache` 强制重新构建 Docker 镜像

**修复命令**:
```bash
docker compose build --no-cache nofx nofx-frontend
docker compose up -d
```

**相关文件**:
- `/root/nofx/api/server.go`
- `/root/nofx/web/src/types.ts`
- `/root/nofx/web/src/components/AITradersPage.tsx`
- `/root/nofx/web/src/components/ModelIcons.tsx`
- `/root/nofx/web/public/icons/gateway.svg`

---

#### BUG-002: 前端 TypeScript 编译错误

**模块**: 前端
**严重程度**: P0
**状态**: ✅ 已修复

**问题描述**:
前端构建时出现 TypeScript 编译错误:
```
src/components/AITradersPage.tsx(1799,64): error TS1005: '}' expected.
```

**原因分析**:
三元表达式嵌套括号不正确，导致 TypeScript 解析错误。

**解决方案**:
修正三元表达式格式:
```tsx
// 修复前 (错误)
{isGateway
  ? (selectedPreset === 'custom'
      ? (language === 'zh' ? '输入自定义 API 中转地址' : 'Enter custom gateway API URL')
      : (language === 'zh' ? '预设地址，可修改' : 'Preset URL, editable'))
    : t('leaveBlankForDefault', language)}

// 修复后 (正确)
{isGateway
  ? selectedPreset === 'custom'
    ? (language === 'zh' ? '输入自定义 API 中转地址' : 'Enter custom gateway API URL')
    : (language === 'zh' ? '预设地址，可修改' : 'Preset URL, editable')
  : t('leaveBlankForDefault', language)}
```

**相关文件**:
- `/root/nofx/web/src/components/AITradersPage.tsx`

---

#### BUG-003: ZenMux Gateway 预设 URL 配置错误

**模块**: 后端
**严重程度**: P1
**状态**: ✅ 已修复

**问题描述**:
在策略工作室选择 "ZenMux.ai" 预设进行 AI 模拟测试时，出现 DNS 解析错误:
```
AI call failed: AI API call failed: still failed after 3 retries: failed to send request:
Post "https://api.zenmux.ai/v1/chat/completions": dial tcp: lookup api.zenmux.ai on 127.0.0.11:53: no such host
```

**原因分析**:
1. `api/server.go` 中 ZenMux 预设 URL 配置错误
2. 错误使用 `https://api.zenmux.ai/v1` 而非正确的 `https://zenmux.ai/api/v1`
3. 域名 `api.zenmux.ai` 不存在（DNS 返回 NXDOMAIN）
4. 正确的域名是 `zenmux.ai`，API 路径为 `/api/v1`
5. **关键问题**：用户之前保存的 AI 模型配置中已存储错误的旧 URL，仅修复代码预设不会影响已保存的配置

**解决方案**:
1. 修正 `api/server.go:2950` 中的预设 URL:
```go
// 修复前
{"id": "zenmux", "name": "ZenMux.ai", "url": "https://api.zenmux.ai/v1"},

// 修复后
{"id": "zenmux", "name": "ZenMux.ai", "url": "https://zenmux.ai/api/v1"},
```

2. 更新数据库中已保存的旧 URL:
```bash
# 备份数据库
docker cp nofx-trading:/app/data/data.db /tmp/data_backup.db

# 复制数据库到本地
docker cp nofx-trading:/app/data/data.db /tmp/data.db

# 使用 Python 更新数据库
python3 << 'EOF'
import sqlite3
conn = sqlite3.connect('/tmp/data.db')
cursor = conn.cursor()
cursor.execute("UPDATE ai_models SET custom_api_url = REPLACE(custom_api_url, 'https://api.zenmux.ai/v1', 'https://zenmux.ai/api/v1') WHERE custom_api_url LIKE '%api.zenmux.ai%'")
conn.commit()
conn.close()
EOF

# 复制回容器并重启
docker cp /tmp/data.db nofx-trading:/app/data/data.db
docker compose restart nofx
```

**临时解决方案**:
如果用户在修复前遇到此问题，可以在配置界面选择 "Custom Gateway" 并手动输入正确的 URL: `https://zenmux.ai/api/v1`，或者直接在前端界面修改已保存模型的 URL。

**修复命令**:
```bash
# 1. 修复代码
# 编辑 api/server.go

# 2. 重新构建后端
docker compose build --no-cache nofx
docker compose up -d

# 3. 更新数据库（见上方 Python 脚本）

# 4. 重启服务
docker compose restart nofx
```

**相关文件**:
- `/root/nofx/api/server.go:2950`
- 数据库表: `ai_models.custom_api_url`

---

#### BUG-004: 回测实验室不支持 Gateway Provider

**模块**: 后端 / 回测
**严重程度**: P1
**状态**: ✅ 已修复

**问题描述**:
在回测实验室选择使用 Gateway (如 ZenMux.ai) 配置的 AI 模型进行回测时，出现错误:
```
unsupported ai provider gateway
```

**原因分析**:
1. `backtest/ai_client.go` 中的 `configureMCPClient` 函数使用 switch 语句处理不同的 AI provider
2. switch 语句中缺少 `gateway` case 的处理
3. Gateway providers (ZenMux, OpenRouter 等) 使用 OpenAI 兼容的 API 格式，应该使用通用的 `mcp.Client` 来处理

**解决方案**:
在 `backtest/ai_client.go` 的 switch 语句中添加 `gateway` case:
```go
case "gateway":
    // Gateway providers (ZenMux, OpenRouter, etc.) use OpenAI-compatible API
    if cfg.AICfg.APIKey == "" {
        return nil, fmt.Errorf("gateway provider requires api key")
    }
    client := mcp.NewClient()
    client.SetAPIKey(cfg.AICfg.APIKey, cfg.AICfg.BaseURL, cfg.AICfg.Model)
    return client, nil
```

**修复命令**:
```bash
# 重新构建后端
docker compose build --no-cache nofx
docker compose up -d
```

**相关文件**:
- `/root/nofx/backtest/ai_client.go:74-81`

---

#### BUG-005: 回测中使用导入策略时出现空指针错误

**模块**: 后端 / 回测
**严重程度**: P1
**状态**: ✅ 已修复
**发现时间**: 2025-12-27
**修复时间**: 2025-12-27

**问题描述**:
在回测功能中使用从本地导入的 JSON 策略进行测试时，出现 Internal Server Error:
```
runtime error: invalid memory address or nil pointer dereference
[GIN] 2025/12/27 - 19:03:48 | 500 | 4.26709079s | POST "/api/backtest/start"
```

**原因分析**:
1. 回测中的 `BuildMarketData` 函数 (`backtest/datafeed.go`) 没有创建 `TimeframeData` 字段
2. 当策略引擎的 `formatMarketData` 函数尝试访问 `data.TimeframeData[tf]` 时，返回 nil 指针
3. 在 `formatTimeframeSeriesData` 函数中访问 `len(data.Klines)` 时触发空指针错误
4. 问题发生在用户导入策略时，策略配置使用了完整的多时间框架数据格式，但回测数据源没有提供相应的 `TimeframeData` 结构

**解决方案**:
在 `backtest/datafeed.go` 的 `BuildMarketData` 函数中添加 `TimeframeData` 字段的构建:
```go
// Build TimeframeData for primary timeframe data to support strategy engine
if primaryData, hasPrimary := perTF[df.primaryTF]; hasPrimary {
    primaryData.TimeframeData = make(map[string]*market.TimeframeSeriesData, len(df.timeframes))
    for _, tf := range df.timeframes {
        // Re-fetch the raw Kline data for this timeframe
        series := df.sliceUpTo(symbol, tf, ts)
        if len(series) == 0 {
            continue
        }

        // Convert Kline to KlineBar for TimeframeSeriesData
        klineBars := make([]market.KlineBar, len(series))
        for i, k := range series {
            klineBars[i] = market.KlineBar{
                Time:   k.OpenTime,
                Open:   k.Open,
                High:   k.High,
                Low:    k.Low,
                Close:  k.Close,
                Volume: k.Volume,
            }
        }

        // Get the calculated indicator data from the built market data
        tfData, hasData := perTF[tf]
        if !hasData || tfData.IntradaySeries == nil {
            continue
        }

        primaryData.TimeframeData[tf] = &market.TimeframeSeriesData{
            Timeframe:   tf,
            Klines:      klineBars,
            MidPrices:   tfData.IntradaySeries.MidPrices,
            EMA20Values: tfData.IntradaySeries.EMA20Values,
            MACDValues:  tfData.IntradaySeries.MACDValues,
            RSI7Values:  tfData.IntradaySeries.RSI7Values,
            RSI14Values: tfData.IntradaySeries.RSI14Values,
            Volume:      tfData.IntradaySeries.Volume,
            ATR14:       tfData.IntradaySeries.ATR14,
        }
    }
}
```

**修复命令**:
```bash
# 重新构建后端
docker compose build nofx
docker compose up -d nofx
```

**相关文件**:
- `/root/nofx/backtest/datafeed.go:141-217`

**测试验证**:
1. 导入一个 JSON 策略文件
2. 在回测实验室中选择该策略
3. 配置回测参数并启动
4. 确认回测正常运行，不再出现空指针错误

---

#### BUG-006: 回测中 DecisionTimeframe 不在 Timeframes 列表时出现空指针错误

**模块**: 后端 / 回测
**严重程度**: P1
**状态**: ✅ 已修复
**发现时间**: 2025-12-27
**修复时间**: 2025-12-27

**问题描述**:
在回测实验室中，当 `DecisionTimeframe` 不在 `Timeframes` 列表中时，出现空指针错误:
```
runtime error: invalid memory address or nil pointer dereference
nofx/backtest/datafeed.go:101 (0xa62860)
```

**原因分析**:
1. 在 `NewDataFeed` 中，`primaryTF` 设置为 `cfg.DecisionTimeframe`
2. 但 `timeframes` 列表直接使用 `cfg.Timeframes`，可能不包含 `DecisionTimeframe`
3. 在 `loadAll()` 中，只加载 `timeframes` 列表中的时间框架数据
4. 当 `DecisionTimeframe` 不在列表中时，`primarySeries` 为 nil，导致空指针错误

**解决方案**:
在 `NewDataFeed` 函数中确保 `DecisionTimeframe` 被包含在要加载的时间框架列表中:
```go
func NewDataFeed(cfg BacktestConfig) (*DataFeed, error) {
    // Ensure DecisionTimeframe is included in the timeframes list
    timeframes := make([]string, 0, len(cfg.Timeframes)+1)
    decisionTFIncluded := false
    for _, tf := range cfg.Timeframes {
        timeframes = append(timeframes, tf)
        if tf == cfg.DecisionTimeframe {
            decisionTFIncluded = true
        }
    }
    // Add DecisionTimeframe if not already in the list
    if !decisionTFIncluded && cfg.DecisionTimeframe != "" {
        timeframes = append(timeframes, cfg.DecisionTimeframe)
    }

    df := &DataFeed{
        cfg:          cfg,
        symbols:      make([]string, len(cfg.Symbols)),
        timeframes:   timeframes,
        symbolSeries: make(map[string]*symbolSeries),
        primaryTF:    cfg.DecisionTimeframe,
    }
    // ...
}
```

**修复命令**:
```bash
# 重新构建后端
docker compose build nofx
docker compose up -d nofx
```

**相关文件**:
- `/root/nofx/backtest/datafeed.go:31-60`

**测试验证**:
1. 在回测实验室中配置任意时间框架组合
2. 确保 DecisionTimeframe 可以不在 Timeframes 列表中
3. 启动回测，确认不再出现空指针错误

---

### 待记录的 Bug

<!--
记录新 Bug 时，请按以下格式添加:

#### BUG-XXX: [Bug 标题]

**模块**: [模块名称]
**严重程度**: [P0/P1/P2/P3]
**状态**: [待修复/修复中/已修复]

**问题描述**:


**原因分析**:


**解决方案**:


**相关文件**:


-->

---

#### BUG-007: 回测中 AI 决策出现 "price unavailable for ALL" 和 "invalid close qty" 错误

**模块**: 后端 / 回测 / AI 模型
**严重程度**: P1
**状态**: ✅ 已修复
**发现时间**: 2025-12-27
**修复时间**: 2025-12-27

**问题描述**:
在回测运行过程中，AI 有时会报以下错误：
1. `price unavailable for ALL` - 当 AI 没有输出有效 JSON 时触发
2. `invalid close qty` - 当 AI 尝试关闭仓位时，明明有持仓却提示数量无效

**原因分析**:

**问题 1: "price unavailable for ALL"**
- 在 `decision/engine.go:1461-1474`，当 AI 没有输出有效 JSON 决策时，系统创建了一个 fallback decision
- 这个 fallback decision 使用 `"ALL"` 作为 symbol，这不是一个有效的交易币种
- 在回测执行时，`priceMap["ALL"]` 不存在，导致 `basePrice <= 0`，触发错误

**问题 2: "invalid close qty"**
- AI 返回的 symbol 格式可能不一致（例如开仓时用 "ETH"，闭仓时用 "ETHUSDT"）
- 在 `account.go:52`，创建仓位时使用 `strings.ToUpper(symbol)`，只转大写不添加 USDT 后缀
- 在 `runner.go:837-844`，`determineCloseQuantity` 函数只用 `strings.ToUpper(symbol)` 查找仓位
- 当 AI 开仓返回 "ETH"（仓位存为 "ETH"），闭仓返回 "ETHUSDT"（查找 "ETHUSDT"）时，找不到匹配的仓位，返回 0，触发错误

**解决方案**:

**修复 1: 移除 fallback decision 中的无效 symbol**
在 `decision/engine.go:1468-1471` 中，将 fallback decision 改为返回空数组：
```go
// 修复前
fallbackDecision := Decision{
    Symbol:    "ALL",  // 无效的交易币种
    Action:    "wait",
    Reasoning: fmt.Sprintf("Model didn't output structured JSON decision..."),
}
return []Decision{fallbackDecision}, nil

// 修复后
// Return empty decisions array instead of a fallback decision with invalid symbol
// This prevents "price unavailable for ALL" errors in backtest
logger.Infof("⚠️  [SafeFallback] Returning empty decisions array - system will wait")
return []Decision{}, nil
```

**修复 2: 使用 market.Normalize 处理 symbol 匹配**
在 `runner.go:837-850` 和 `account.go:224-240` 中，添加 symbol normalize 逻辑：

`runner.go` 修复：
```go
func (r *Runner) determineCloseQuantity(symbol, side string, dec decision.Decision) float64 {
    // Normalize symbol for consistent comparison (handles both "ETH" and "ETHUSDT" formats)
    normalizedSymbol := market.Normalize(symbol)
    upperSymbol := strings.ToUpper(symbol)

    for _, pos := range r.account.Positions() {
        // Try both normalized and uppercase formats for matching
        // This handles cases where AI returns inconsistent symbol formats
        if (pos.Symbol == normalizedSymbol || pos.Symbol == upperSymbol) && pos.Side == side {
            return pos.Quantity
        }
    }
    return 0
}
```

`account.go` 修复：
```go
func (acc *BacktestAccount) positionLeverage(symbol, side string) int {
    // Try both normalized and uppercase formats to handle AI symbol format inconsistency
    key := positionKey(symbol, side)
    if pos, ok := acc.positions[key]; ok && pos.Quantity > epsilon {
        return pos.Leverage
    }

    // Try with normalized symbol (e.g., "ETH" -> "ETHUSDT")
    normalizedSymbol := market.Normalize(symbol)
    normalizedKey := positionKey(normalizedSymbol, side)
    if pos, ok := acc.positions[normalizedKey]; ok && pos.Quantity > epsilon {
        return pos.Leverage
    }

    return 0
}
```

**修复命令**:
```bash
# 重新构建后端
docker compose build nofx
docker compose up -d nofx
```

**相关文件**:
- `/root/nofx/decision/engine.go:1468-1471`
- `/root/nofx/backtest/runner.go:837-850`
- `/root/nofx/backtest/account.go:3-9, 224-240`

**测试验证**:
1. 运行回测，观察 AI 决策日志
2. 当 AI 没有输出有效 JSON 时，确认不再出现 "price unavailable for ALL" 错误
3. 当 AI 开仓和闭仓使用不同格式的 symbol 时（如 "ETH" vs "ETHUSDT"），确认正常关闭仓位

**实盘交易影响**:
- 实盘交易不受此问题影响
- `market.Get()` 函数会自动 normalize symbol（`market/data.go:103`）
- 交易所 API 会处理不存在仓位的情况

---

### 待记录的 Bug

<!--
记录新 Bug 时，请按以下格式添加:

#### BUG-XXX: [Bug 标题]

**模块**: [模块名称]
**严重程度**: [P0/P1/P2/P3]
**状态**: [待修复/修复中/已修复]

**问题描述**:


**原因分析**:


**解决方案**:


**相关文件**:


-->

---

## 修复流程

### Bug 报告流程

1. **发现 Bug**
   - 记录 Bug 现象
   - 确定严重程度和模块分类
   - 收集相关日志和截图

2. **创建 Bug 记录**
   - 在本文档中添加 Bug 记录
   - 分配 Bug ID (格式: BUG-XXX)
   - 填写详细信息

3. **分析和修复**
   - 定位问题根因
   - 设计解决方案
   - 编写修复代码

4. **测试验证**
   - 本地测试修复效果
   - 确认没有引入新问题
   - 更新文档

5. **部署上线**
   - 合并代码到主分支
   - 重新构建并部署
   - 验证线上环境

### Bug 修复模板

```markdown
#### BUG-XXX: [Bug 标题]

**模块**: [模块名称]
**严重程度**: [P0/P1/P2/P3]
**状态**: [待修复/修复中/已修复]
**发现时间**: [YYYY-MM-DD]
**修复时间**: [YYYY-MM-DD]

**问题描述**:


**原因分析**:


**解决方案**:


**相关文件**:


**测试验证**:


```

---

## 统计信息

| 统计项 | 数量 |
|--------|------|
| 总 Bug 数 | 7 |
| 已修复 | 7 |
| 待修复 | 0 |
| P0 级别 | 1 |
| P1 级别 | 6 |
| P2 级别 | 0 |
| P3 级别 | 0 |

---

## 联系方式

如有 Bug 需要报告，请联系:
- **项目 Issues**: [GitHub Issues](https://github.com/your-org/nofx/issues)
- **开发者社区**: [NOFX Developer Community](https://t.me/nofx_dev_community)

---

*最后更新: 2025-12-27*
