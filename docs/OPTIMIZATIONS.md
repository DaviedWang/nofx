# NOFX 产品优化文档

> 本文档记录 NOFX 项目的所有产品功能优化、改进和新特性，方便团队跟踪产品演进历程。

---

## 目录

1. [优化分类说明](#优化分类说明)
2. [优化记录](#优化记录)
3. [规划中的优化](#规划中的优化)

---

## 优化分类说明

### 优化类型

| 类型 | 说明 | 图标 |
|------|------|------|
| **新功能** | 全新的功能特性 | ✨ |
| **增强** | 现有功能的改进 | 🔧 |
| **性能** | 性能优化 | ⚡ |
| **体验** | 用户体验优化 | 💎 |
| **修复** | Bug 修复 | 🐛 |
| **重构** | 代码重构 | 🔄 |

### 模块分类

- **AI 模型 (AI Model)**: AI 模型相关功能
- **交易所 (Exchange)**: 交易所相关功能
- **界面 (UI/UX)**: 用户界面优化
- **策略 (Strategy)**: 交易策略相关
- **系统 (System)**: 系统级改进

---

## 优化记录

### 2025-12-27

---

#### OPT-006: 默认保证金使用率调整为 50%

**类型**: 🔧 增强
**模块**: 策略
**状态**: ✅ 已完成

**功能描述**:
将默认的最大保证金使用率从 90% 调整为 50%，使系统默认更加保守，降低风险暴露。

**修改内容**:
1. `backtest/config.go:279` - MaxMarginUsage 从 0.9 改为 0.5
2. `backtest/runner.go:762` - 默认值从 0.9 改为 0.5

**效果**:
- 所有新建回测的默认最大保证金使用率为 50%
- 单个 ETHUSDT 仓位最大约为 2500 USDT（而不是之前的 4500 USDT）
- AI prompt 中也会显示 "Max allowed: 50%"

**相关文件**:
- `/root/nofx/backtest/config.go:279`
- `/root/nofx/backtest/runner.go:759-763`

---

#### OPT-005: AI 仓位建议保守化 - 防止过度激进开仓

**类型**: 🔧 增强
**模块**: 策略
**状态**: ✅ 已完成

**功能描述**:
优化 AI 决策 prompt 中的仓位建议，从激进的建议（80-100% 最大限制）改为保守的建议（40-60%），并添加当前保证金使用率的实时警告。

**问题背景**:
用户报告回测中 AI 开出了 4500 USDT 的 ETHUSDT 仓位（账户权益仅 997 USDT），保证金使用率高达 90.3%。日志显示：
```
🛑 RISK CONTROL: Capping position from 10000.00 to 4500.00
```
AI 请求了 10000 USDT（10倍于权益），虽然被风控截断到 4500 USDT，但问题在于 AI 的建议本身过于激进。

**问题原因**:
1. 原 prompt 中建议："High confidence (≥85): Use 80-100% of max position value limit"
2. AI 理解为应该使用接近上限的仓位，所以请求了 10000 USDT
3. Prompt 没有显示当前保证金使用率，AI 不知道账户风险状况
4. 40-60% 的建议更符合实际风险控制需求

**优化内容**:

1. **修改仓位建议比例**
   ```go
   // 修改前 (激进)
   - High confidence (≥85): Use 80-100% of max position value limit
   - Medium confidence (70-84): Use 50-80% of max position value limit
   - Low confidence (60-69): Use 30-50% of max position value limit

   // 修改后 (保守)
   - High confidence (≥85): Use 40-60% of max position value limit (conservative)
   - Medium confidence (70-84): Use 25-40% of max position value limit
   - Low confidence (60-69): Use 15-25% of max position value limit
   ```

2. **添加实时保证金使用率显示**
   ```go
   sb.WriteString("**IMPORTANT**: Be conservative with position sizing. Consider margin usage!\n")
   sb.WriteString(fmt.Sprintf("Current margin usage: %.1f%% | Max allowed: %.0f%%\n",
       marginUsedPct, riskControl.MaxMarginUsage*100))
   if marginUsedPct > 50 {
       sb.WriteString(fmt.Sprintf("⚠️ WARNING: Margin usage %.1f%% is already high! Use smaller positions.\n", marginUsedPct))
   }
   ```

3. **添加明确的保守提示**
   ```go
   sb.WriteString("- **DO NOT** use 100% of max position value limit - stay conservative!\n\n")
   ```

4. **更新函数签名传递保证金使用率**
   ```go
   func (e *StrategyEngine) BuildSystemPrompt(accountEquity float64, variant string, marginUsedPct float64) string
   ```

**预期效果**:
- AI 将请求更保守的仓位大小（如 1800-2700 USDT 而不是 10000 USDT）
- 当保证金使用率 >50% 时，AI 会看到警告并进一步减小仓位
- 减少风控截断的频率，让 AI 主动遵守风险限制

**技术实现**:
- 修改 `decision/engine.go` 中的 `BuildSystemPrompt` 函数
- 添加 `marginUsedPct` 参数并实时显示当前保证金使用率
- 更新所有调用点传递保证金使用率参数
- 在 `decision/engine.go:257` 中使用 `ctx.Account.MarginUsedPct`

**相关文件**:
- `/root/nofx/decision/engine.go:257` (调用点)
- `/root/nofx/decision/engine.go:725` (函数定义)
- `/root/nofx/decision/engine.go:782-800` (仓位建议部分)
- `/root/nofx/api/strategy.go:359-363` (API 调用点)
- `/root/nofx/api/strategy.go:487` (测试调用点)
- `/root/nofx/debate/engine.go:183` (辩论调用点)
- `/root/nofx/debate/engine.go:546` (投票调用点)

**测试验证**:
1. 运行回测，观察 AI 请求的仓位大小
2. 检查日志中的风控截断信息是否减少
3. 确认 AI 在高保证金使用率时主动减小仓位

---

#### OPT-004: 回测风控严格化 - 严格执行保证金限制

**类型**: 🔧 增强
**模块**: 策略
**状态**: ✅ 已完成

**功能描述**:
强化回测模块的风控逻辑，确保 AI 严格遵守策略配置中设置的保证金限制，防止违规开仓导致保证金使用率超标。

**问题背景**:
用户报告在回测中发现，虽然策略风控配置中设置了最大保证金使用率（如 80%），但 AI 仍然违规开仓，导致保证金使用率达到 90%。这说明原有的风控逻辑存在缺陷，没有正确执行策略配置中的限制。

**问题原因**:
1. 原 `determineQuantity` 函数只考虑可用现金（availableCash），没有计算当前已使用的保证金总额
2. 没有使用策略配置中的 `MaxMarginUsage` 限制参数
3. 没有检查单个仓位的最大价值限制（`BTCETHMaxPositionValueRatio` / `AltcoinMaxPositionValueRatio`）
4. 使用固定的 90% 限制，忽略了策略配置

**优化内容**:

1. **保证金总额检查**
   ```go
   // Calculate current margin usage
   currentMarginUsed := r.totalMarginUsed()
   maxAllowedMargin := equity * maxMarginUsage
   remainingMargin := maxAllowedMargin - currentMarginUsed
   ```

2. **使用策略配置的 MaxMarginUsage**
   ```go
   // Use strategy's MaxMarginUsage if set (default to 0.9 if not set)
   maxMarginUsage := riskControl.MaxMarginUsage
   if maxMarginUsage <= 0 || maxMarginUsage > 1 {
       maxMarginUsage = 0.9
   }
   ```

3. **单个仓位价值限制**
   ```go
   // Check single position value limit
   if sym == "BTCUSDT" || sym == "ETHUSDT" {
       maxPositionValueRatio = riskControl.BTCETHMaxPositionValueRatio
       if maxPositionValueRatio <= 0 {
           maxPositionValueRatio = 5.0 // Default: 5x equity
       }
   } else {
       maxPositionValueRatio = riskControl.AltcoinMaxPositionValueRatio
       if maxPositionValueRatio <= 0 {
           maxPositionValueRatio = 1.0 // Default: 1x equity
       }
   }
   maxSinglePositionValue := equity * maxPositionValueRatio
   ```

4. **取最严格的限制**
   ```go
   // Use the more restrictive limit
   if maxPositionValue > maxSinglePositionValue {
       maxPositionValue = maxSinglePositionValue
   }
   ```

5. **详细的日志输出**
   ```go
   logger.Infof("🛑 RISK CONTROL: Capping position from %.2f to %.2f (max allowed: %.2f, reason: %s)",
       sizeUSD, maxPositionValue, maxPositionValue, reason)
   ```

**风控检查流程**:
```
1. 获取策略配置中的风控参数
   ↓
2. 计算当前已使用的保证金总额
   ↓
3. 根据权益 × MaxMarginUsage 计算最大允许保证金
   ↓
4. 计算剩余可用保证金
   ↓
5. 根据剩余保证金 × 杠杆计算最大仓位价值
   ↓
6. 检查单个仓位价值限制（BTC/ETH vs Altcoin）
   ↓
7. 取两个限制中的较小值
   ↓
8. 如果 AI 请求的仓位超过限制，截断并记录日志
   ↓
9. 如果剩余保证金不足，拒绝开仓
```

**用户价值**:
- 确保回测结果与实盘风控行为一致
- 防止 AI 违规开仓导致的过度风险暴露
- 提供清晰的风控日志，便于调试和优化
- 支持灵活的保证金使用率配置

**技术实现**:
- 修改 `/root/nofx/backtest/runner.go` 中的 `determineQuantity` 函数
- 使用 `strategyEngine.GetConfig()` 获取策略配置
- 添加保证金总额计算和剩余保证金检查
- 添加单个仓位价值限制检查
- 增强日志输出，明确标识风控触发原因

**相关文件**:
- `/root/nofx/backtest/runner.go:742-835`

**测试验证**:
1. 设置策略风控配置：MaxMarginUsage = 0.6 (60%)
2. 运行回测，观察保证金使用率是否被严格限制在 60% 以下
3. 查看日志中的 "🛑 RISK CONTROL" 标记，确认风控正确触发

---

### 2025-12-26

---

#### OPT-001: 新增 API Gateway 中转服务支持

**类型**: ✨ 新功能
**模块**: AI 模型
**状态**: ✅ 已完成

**功能描述**:
新增 API Gateway (中转服务) 支持，允许用户通过统一的中转服务访问多个 AI 模型，方便模型切换和成本控制。

**优化内容**:

1. **预设中转服务支持**
   - ZenMux.ai
   - OpenRouter.ai
   - 自定义中转服务

2. **配置界面优化**
   - 新增中转服务选择下拉框
   - API URL 自动填充预设地址
   - 支持自定义 API 地址

3. **模型灵活切换**
   - 通过"自定义模型名称"字段自由切换模型
   - 支持 gpt-4o-mini, gpt-4o, claude-3-5-sonnet-20241022 等多种模型

**用户价值**:
- 降低模型切换成本，无需配置多个 API Key
- 支持更多中转服务，方便选择性价比方案
- 简化配置流程，提高易用性

**技术实现**:
- 后端 `api/server.go` 添加 gateway provider
- 前端 `types.ts` 新增 `GatewayPreset` 类型
- 前端 `ModelConfigModal` 组件新增 gateway 配置界面
- 创建 `gateway.svg` 图标

**使用方式**:
1. 访问 AI 模型配置页面
2. 选择 "API Gateway (中转服务)"
3. 选择预设中转服务或输入自定义 API 地址
4. 输入 API Key
5. 在"自定义模型名称"中输入目标模型 ID

**相关文件**:
- `/root/nofx/api/server.go`
- `/root/nofx/web/src/types.ts`
- `/root/nofx/web/src/components/AITradersPage.tsx`
- `/root/nofx/web/src/components/ModelIcons.tsx`
- `/root/nofx/web/public/icons/gateway.svg`

---

#### OPT-002: 前端类型定义增强

**类型**: 🔧 增强
**模块**: 界面
**状态**: ✅ 已完成

**功能描述**:
增强前端 TypeScript 类型定义，为 API Gateway 功能提供类型支持。

**优化内容**:
- `AIModel` 接口新增 `description` 字段
- `AIModel` 接口新增 `presets` 字段
- `AIModel` 接口新增 `selectedPreset` 字段
- 新增 `GatewayPreset` 接口定义

**相关文件**:
- `/root/nofx/web/src/types.ts`

---

#### OPT-003: AI 模型配置界面优化

**类型**: 💎 体验
**模块**: 界面
**状态**: ✅ 已完成

**功能描述**:
优化 AI 模型配置弹窗，针对 Gateway 类型显示专门的配置界面。

**优化内容**:
1. **动态表单**: 根据选择的模型类型显示不同配置项
2. **预设选择**: Gateway 类型显示中转服务下拉选择
3. **智能提示**: 根据配置状态显示不同的帮助文本
4. **国际化支持**: 中英文双语提示

**用户体验改进**:
- 减少配置步骤，预设服务自动填充 URL
- 清晰的提示信息，降低配置错误
- 视觉反馈，Gateway 使用金色图标区分

**相关文件**:
- `/root/nofx/web/src/components/AITradersPage.tsx`

---

### 历史优化

<!--
记录历史优化时，请按以下格式添加:

#### OPT-XXX: [优化标题]

**类型**: [新功能/增强/性能/体验/修复/重构]
**模块**: [模块名称]
**状态**: [已完成/进行中/计划中]
**日期**: [YYYY-MM-DD]

**功能描述**:


**优化内容**:


**用户价值**:


**相关文件**:


-->

---

## 规划中的优化

### 待规划的优化

<!--
规划新优化时，请按以下格式添加:

#### OPT-XXX: [优化标题]

**类型**: [新功能/增强/性能/体验/修复/重构]
**模块**: [模块名称]
**状态**: [计划中/评审中/开发中]
**预计完成**: [YYYY-MM-DD]

**功能描述**:


**目标**:


**优先级**: [P0/P1/P2/P3]


-->

---

## 统计信息

| 统计项 | 数量 |
|--------|------|
| 总优化数 | 6 |
| 已完成 | 6 |
| 进行中 | 0 |
| 计划中 | 0 |
| 新功能 | 1 |
| 增强 | 4 |
| 体验 | 1 |

---

## 优化提案

如需提出新的优化建议，请按以下模板提交:

```markdown
#### 优化提案: [标题]

**类型**: [新功能/增强/性能/体验/修复/重构]
**模块**: [模块名称]
**优先级**: [P0/P1/P2/P3]

**当前问题**:


**优化建议**:


**预期收益**:


**实现复杂度**: [低/中/高]


```

---

## 联系方式

如有优化建议或需求，请联系:
- **项目 Issues**: [GitHub Issues](https://github.com/your-org/nofx/issues)
- **核心团队**: Tinkle (@Web3Tinkle)
- **官方 Twitter**: [@nofx_official](https://x.com/nofx_official)

---

*最后更新: 2025-12-27*
