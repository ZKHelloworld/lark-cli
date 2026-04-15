# lark-doc 画板处理指南

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

## docs +whiteboard-update 说明

`docs +whiteboard-update` 是 `whiteboard +update` 在 `docs` 域下的别名，功能完全相同。推荐优先使用 `whiteboard +update`；已有调用使用 `docs +whiteboard-update` 继续有效。

## 三个 Skill 的职责边界

| Skill | 核心职责 | 何时使用 |
|------|------|------|
| `lark-doc` | 文档内容读取/更新、插入空白画板占位、拿到 board token | 创建或编辑文档时需要嵌入画板 |
| `lark-whiteboard` | 查询与更新已有画板（基于 board token） | 已有 token，需要读图或改图 |
| `lark-whiteboard-cli` | 图表内容生成全流程 — Mermaid/DSL 路由、场景选型、DSL 生成与渲染验证 | 需要生成图表内容时（推荐作为标准图表生成入口） |

## 路由规则

1. 需要在文档中放一个新画板：先用 `docs +create` 或 `docs +update` 插入空白画板。
2. 需要更新已有画板内容：先通过 `docs +fetch` 或 `docs +create`/`+update` 响应获取 board_token，再用 `whiteboard-cli` 生成新内容，最后用 `whiteboard +update --overwrite` 写入。
3. 需要绘制图表内容：先用 [`../../lark-whiteboard-cli/SKILL.md`](../../lark-whiteboard-cli/SKILL.md) 产出 DSL/Mermaid，再用 whiteboard 写入。

## 文档与画板协同流程（完整执行）

### 步骤 1：在文档中创建空白画板

- 创建场景：`docs +create`
- 编辑场景：`docs +update`
- 在 markdown 中使用 `<whiteboard type="blank"></whiteboard>`（不要转义）
- 需要多个画板时，在同一段 markdown 中重复多个 whiteboard 标签

### 步骤 2：获取 board token

- 从 `docs +update` 或 `docs +create` 的响应中读取 `data.board_tokens`
- 如果是读取已有文档，可先 `docs +fetch`，再从 `<whiteboard token="xxx"/>` 解析 token

### 步骤 3：委派 subAgent 生成并写入画板内容

**每个 board_token 启动一个独立 subAgent**，主 agent 不直接执行画板生成管线。

按下方"subAgent 委派规范"组装 5 项输入，并行（或顺序）调用 subAgent，汇总结果。

### 步骤 4：完成校验

- 确认每个 token 对应的画板都已填充真实内容
- 不保留空白占位画板

## 语义与画板类型映射

| 语义 | 画板类型 |
|------|------|
| 架构/分层/技术方案/模块依赖/调用关系 | 架构图 |
| 流程/审批/部署/业务流转/状态机 | 流程图 |
| 跨角色流程/跨系统交互/端到端链路 | 泳道图 |
| 组织/层级/汇报关系 | 组织架构图 |
| 时间线/里程碑/版本规划 | 里程碑图 |
| 因果/复盘/根因分析 | 鱼骨图 |
| 方案对比/技术选型/功能矩阵 | 对比图 |
| 循环/飞轮/闭环/增长链路 | 飞轮图 |
| 层级占比/能力模型/需求层次 | 金字塔图 |
| 矩形树图/层级面积占比 | 树状图 |
| 转化漏斗/销售漏斗 | 漏斗图 |
| 分类梳理/知识体系/思维导图/时序图/类图 | Mermaid |
| 数据分布/占比/饼图 | Mermaid |
| 柱状图/条形图/数据对比 | 柱状图 |
| 折线图/趋势图/时序数据 | 折线图 |

## 注意事项

- `lark-doc` 不能直接编辑已有画板内容；它负责文档层面的插入与 token 获取。
- `docs +update` 的画板能力是插入空白占位，不是画板内容编辑器。
- 只创建空白画板而不填充内容，任务视为未完成。

## 关联参考

- 画板内容生成：[`../../lark-whiteboard-cli/SKILL.md`](../../lark-whiteboard-cli/SKILL.md)
- 画板读写与更新：[`../../lark-whiteboard/SKILL.md`](../../lark-whiteboard/SKILL.md)

---

## subAgent 委派规范

> **⚠️ CRITICAL：主 Agent 越权禁止令**
> **主 Agent 绝对禁止**自行编写任何图表代码（哪怕是最简单的 Mermaid 流程图）并直接调用 `whiteboard +update`。
> - **强制委派**：必须 100% 将 `board_token` 委派给 `lark-whiteboard-cli` subAgent 处理。跳过 subAgent 直接修改画板将被视为严重违规和任务失败。
> - **为什么不能自己写？** 因为 `lark-whiteboard-cli` 技能内封装了专属的 DSL 转换引擎、防截断渲染校验以及特定场景模板库。主 Agent 绕过它直接写入会导致图表缺乏校验、样式错乱甚至引发系统渲染错误。

### 输入契约（5 项，缺一不可）

主 agent 为**每个** board_token 单独组装以下 5 项输入：

| # | 字段 | 内容要求 |
|---|------|---------|
| 1 | **画板编号** | 整数（从 1 开始），用于产物目录命名（`./diagrams/board_{n}/`） |
| 2 | **文档背景** | 1–3 句话，说明文档主题、目标读者、场景 |
| 3 | **画板内容** | 图表类型 + 关键元素 + 元素间关系 + 具体文字/数据（越具体越好） |
| 4 | **board_token** | 目标画板的 token 值（主 agent 从 `data.board_tokens` 取得） |
| 5 | **skill 路径指引** | 明确告知读取 `skills/lark-whiteboard-cli/SKILL.md` 并按其 Workflow 执行 |

> **为什么不需要传完整上传命令？** `lark-whiteboard-cli/SKILL.md` 已内置完整的写入流程（含 dry-run 安全检查和 `lark-cli whiteboard +update` 调用）。subAgent 读完 SKILL.md、生成内容后，按 SKILL.md 的"上传飞书画板"章节执行即可，不需要主 agent 额外构造命令。

### subAgent Prompt 模板

```
你是一个专注于飞书画板内容生成与写入的 agent，使用最强可用模型执行。

## 任务

为第 {{board_index}} 号画板生成并写入图表内容。

## 文档背景

{{doc_background}}

## 画板内容要求

- 图表类型：{{chart_type}}
- 关键元素：{{key_elements}}
- 元素间关系：{{relationships}}
- 具体文字/数据：{{text_data}}

## 目标画板

board_token：{{board_token}}

## 执行指引

1. 读取 `skills/lark-whiteboard-cli/SKILL.md`，按其完整 Workflow 执行
2. 产物目录使用 `./diagrams/board_{{board_index}}/`（固定编号，非时间戳）
3. 生成内容后按 SKILL.md 的"上传飞书画板"章节将内容写入上方 board_token
4. 完成后返回：`{ "board_token": "{{board_token}}", "status": "ok" }`
   若失败：`{ "board_token": "{{board_token}}", "status": "failed", "error": "原因" }`

## 约束

- 只操作画板，不修改文档主体
- 渲染 2 轮后仍有严重问题 → 记录 error，不要无限重试
- 不得在未写入真实内容的情况下返回 ok
```

### 调度策略

| 画板数量 | 调度方式 |
|----------|---------|
| 1 个 | 直接调用单个 subAgent |
| 多个 + 框架支持并行 | 同时发起所有 subAgent（推荐） |
| 多个 + 框架不支持并行 | 按编号顺序串行调用 |

**跨框架适配：**

```
Claude Code   → Task tool（Agent mode）
Copilot Agent → runSubagent
Cursor        → parallel tool calls 或 sequential subagent
其他框架       → 使用平台提供的子任务/并发能力
```

### 主 agent 汇总逻辑

```
所有 subAgent 完成后：
  全部 ok    → 告知用户文档与画板均已完成
  部分 failed → 展示失败的 board_token 和原因，建议用户补充更多细节后重试
  注意：即使全部失败，文档本体（含空白画板占位）已创建，需如实告知用户
```