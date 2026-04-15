# 画板执行 Subagent 化方案

## 设计原则

| 原则 | 说明 |
|------|------|
| **必须用 subAgent** | 画板生成流程长（读指南 → 生成 DSL → 渲染 PNG → 质量审视 → 迭代修正），独立 subAgent 保证每张图有完整注意力，避免上下文污染 |
| **一图一 subAgent** | 每个 board_token 对应一个独立 subAgent，不共享上下文，可并行执行 |
| **编号隔离产物目录** | 多画板时 subAgent 使用 `./diagrams/board_{n}/` 目录（非时间戳），产物路径可预期且不冲突 |
| **用最强可用模型** | subAgent 调用时指定最强可用模型，画板质量优先 |
| **跨框架适配** | Claude Code 用 Agent tool；其他框架用对应子任务能力（sub-agent、parallel tool calls 等） |

---

## Skill 职责边界（读完所有文件后的整理）

通读三个 skill 的所有文件后，职责边界如下：

```
lark-doc
  职责：文档层操作
  ├─ 创建/更新文档内容（Markdown）
  ├─ 在文档中插入空白画板占位（<whiteboard type="blank">）
  └─ 从响应 data.board_tokens 中提取 token

lark-whiteboard
  职责：画板 API 读写层（薄封装）
  ├─ +query：导出画板为 image / code(Mermaid|PlantUML) / raw JSON
  └─ +update：写入画板内容，支持 plantuml / mermaid / raw(DSL JSON)
              必须通过 --source - 或 --source @file 传入内容

lark-whiteboard-cli
  职责：图表内容生成全流程（重逻辑）
  ├─ Step 1：路由（Mermaid vs DSL）+ 读 scene 指南 + 读 6 个核心参考文件 
  ├─ Step 2：生成完整 DSL/Mermaid，产物写入 ./diagrams/<目录>/
  ├─ Step 3：渲染 PNG → 质量审视 → 最多 2 轮修复
  └─ 上传：内置 token 获取 + MANDATORY dry-run 检查 + 实际写入命令
           dry-run是针对已有内容的画板的覆盖前确认，新建的空白画板可跳过
```

**关键发现：**
- `lark-whiteboard-cli` 已内置完整的上传流程（包含 dry-run 安全检查），subAgent 读完 SKILL.md 后自行执行即可，主 agent **不需要**单独传入完整 upload 命令——只传 board_token 即可
- `lark-whiteboard +update` 是统一的写入 API，whiteboard-cli 的上传章节通过管道调用它
- 文件产物默认按时间戳命名目录（`./diagrams/YYYY-MM-DDTHHMMSS/`）；**多 subAgent 并行时需改为按编号命名**（`./diagrams/board_{n}/`），避免时间戳相同时冲突

---

## 执行流对比

**改前（主 agent 全包）：**

```
lark-doc（主 agent）
  ├─ 读 references/lark-doc-whiteboard.md
  ├─ 读 lark-whiteboard-cli/SKILL.md
  │    ├─ 读 references/schema.md + layout.md + style.md
  │    ├─ 读 references/connectors.md + content.md + typography.md (6个)
  │    ├─ 读 scenes/<类型>.md
  │    └─ 写脚本 → node → 渲染 PNG → 检查 → 最多 2 轮修复（串行，阻塞）
  └─ 读 lark-whiteboard/SKILL.md → 调用 +update 写入 board_token（串行）
```

**改后（委派 subAgent，可并行）：**

```
lark-doc（主 agent）
  ├─ 创建/更新文档，插入 <whiteboard> 占位
  ├─ 提取 board_tokens（假设有 N 个）
  ├─ 为每个 token 组装 subAgent 输入（5 项）
  ├─ 并行调用 N 个 subAgent（每个独立上下文，使用最强模型）
  └─ 汇总各 subAgent 结果 → 告知用户

画板 subAgent（每个 board_token 一个，独立上下文）
  ├─ 自行读取 lark-whiteboard-cli/SKILL.md 并按其 Workflow 执行
  ├─ 使用 ./diagrams/board_{n}/ 目录存放产物（非时间戳）
  ├─ 生成 DSL/Mermaid → 渲染 PNG → 质量审视 → 最多 2 轮
  ├─ 执行 whiteboard-cli 内置上传流程将内容写入画板 token
  └─ 报告结果（ok / failed + 原因）
```

---

## subAgent 输入契约（5 项，缺一不可）

主 agent 为**每个** board_token 单独组装以下 5 项输入：

| # | 字段 | 内容要求 |
|---|------|---------|
| 1 | **画板编号** | 整数（从 1 开始），用于产物目录命名（`./diagrams/board_{n}/`） |
| 2 | **文档背景** | 1–3 句话，说明文档主题、目标读者、场景 |
| 3 | **画板内容** | 图表类型 + 关键元素 + 元素间关系 + 具体文字/数据（越具体越好） |
| 4 | **board_token** | 目标画板的 token 值（主 agent 从 `data.board_tokens` 取得） |
| 5 | **skill 路径指引** | 明确告知读取 `skills/lark-whiteboard-cli/SKILL.md` 并按其 Workflow 执行 |

> **为什么不需要传完整上传命令？** `lark-whiteboard-cli/SKILL.md` 已内置完整的写入流程（含 dry-run 安全检查和 `lark-cli whiteboard +update` 调用）。subAgent 读完 SKILL.md、生成内容后，按 SKILL.md 的"上传飞书画板"章节执行即可，不需要主 agent 额外构造命令。

---

## subAgent Prompt 模板

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

---

## 主 agent 的调度策略

### 何时并行

| 画板数量 | 调度方式 |
|----------|---------|
| 1 个 | 直接调用单个 subAgent |
| 多个 + 框架支持并行 | 同时发起所有 subAgent（推荐） |
| 多个 + 框架不支持并行 | 按编号顺序串行调用 |

### 框架适配

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

---

## 需要修改的文件

### 1. `skills/lark-doc/SKILL.md` — 绘图引导第 3 步

SKILL.md 只保留指针，不承载调度逻辑。所有约束集中在 `lark-doc-whiteboard.md`，避免两个文件产生矛盾。

**改前：**
```markdown
### 3. 按顺序阅读并执行

1. `references/lark-doc-whiteboard.md`：绘图路由、三 skill 分工边界、协同流程与图表类型映射
2. `../lark-whiteboard-cli/SKILL.md`：图表内容生成完整流程 — Mermaid/DSL 路由、场景选型、DSL 生成与渲染验证
3. `../lark-whiteboard/SKILL.md`：通过 board_token 将生成的内容写入画板、查询与导出
```

**改后：**
```markdown
### 3. 执行画板

读取 [`references/lark-doc-whiteboard.md`](references/lark-doc-whiteboard.md) 并按其流程执行（包含 subAgent 委派规范）。
```

---

### 2. `skills/lark-doc/references/lark-doc-whiteboard.md` — 步骤 3 改写 + 新增委派规范

**步骤 3 改前：**
```markdown
### 步骤 3：生成与更新画板内容

1. 参考 `../../lark-whiteboard-cli/SKILL.md` 完成图表内容生成（Mermaid/DSL 路由、场景选型、DSL 生成与渲染验证）。
2. 使用 `../../lark-whiteboard/SKILL.md` 的 `+update` 将生成内容写入指定 token。
```

**步骤 3 改后：**
```markdown
### 步骤 3：委派 subAgent 生成并写入画板内容

**每个 board_token 启动一个独立 subAgent**，主 agent 不直接执行画板生成管线。

按下方"subAgent 委派规范"组装 5 项输入，并行（或顺序）调用 subAgent，汇总结果。
```

**在文件末尾新增"subAgent 委派规范"节**，包含：
- 5 项输入契约表格
- subAgent Prompt 模板（含 `./diagrams/board_{n}/` 目录约束）
- 调度策略（并行/串行规则、跨框架适配）
- 主 agent 汇总逻辑

---

### 3. `skills/lark-whiteboard-cli/SKILL.md` — Workflow 前新增 subAgent 目录约束

**⚠️ 该文件不能直接编辑**，它由独立源码仓库构建生成，直接改会在下次同步时被覆盖。

**实现路径：**

```bash
# 1. 在源码仓库修改对应文件
cd ~/Projects/whiteboard-docx-block/packages/whiteboard-skill
# 编辑 SKILL.md 源文件（路径以该仓库实际结构为准）

# 2. 构建
pnpm build

# 3. 同步到本仓库
cp -r ./dist/  ~/Projects/larksuite-cli/skills/lark-whiteboard-cli/
# （具体产物路径以 build 输出为准）
```

**变更内容**：在源文件的 `## Workflow` 节开头新增：

```markdown
> **当被作为 subAgent 调用时**：产物目录使用 `./diagrams/board_{n}/`（n 为 prompt 中传入的画板编号），
> 而非时间戳目录，保证多个并行 subAgent 的产物互不冲突。
```

这是对现有文件产物规范的**补充约束**，不需要改动其他内容。

---

## 收益

| 维度 | 改前 | 改后 |
|------|------|------|
| 主 agent 加载文件数 | ~10（含所有画板参考模块） | ~2（lark-doc-whiteboard.md + lark-shared） |
| 画板失败影响范围 | 阻塞整个文档工作流 | 仅单个 subAgent 报错，文档本体已创建 |
| 多画板处理 | 串行 | 并行（各 subAgent 独立上下文） |
| 产物目录冲突 | 时间戳相同时可能冲突 | board 编号唯一，不冲突 |
| 重试粒度 | 重跑整个文档流程 | 只重跑失败的单个 subAgent |
| 质量保证 | 主 agent 注意力分散 | 每图独占完整注意力 + 最强模型 |
| 上传安全 | 同前 | 同前（whiteboard-cli 内置 dry-run 检查，subAgent 遵循） |
