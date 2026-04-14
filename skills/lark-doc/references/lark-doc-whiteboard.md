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

### 步骤 3：生成与更新画板内容

1. 参考 [`../../lark-whiteboard-cli/SKILL.md`](../../lark-whiteboard-cli/SKILL.md) 完成图表内容生成（Mermaid/DSL 路由、场景选型、DSL 生成与渲染验证）。
2. 使用 [`../../lark-whiteboard/SKILL.md`](../../lark-whiteboard/SKILL.md) 的 `+update` 将生成内容写入指定 token。

### 步骤 4：完成校验

- 确认每个 token 对应的画板都已填充真实内容
- 不保留空白占位画板

## 语义与画板类型映射

| 语义 | 画板类型 | 参考指南 |
|------|------|------|
| 架构/分层/技术方案 | 架构图 | [../../lark-whiteboard-cli/scenes/architecture.md](../../lark-whiteboard-cli/scenes/architecture.md) |
| 流程/审批/部署/业务流转 | 流程图 | [../../lark-whiteboard-cli/scenes/flowchart.md](../../lark-whiteboard-cli/scenes/flowchart.md) |
| 组织/层级/汇报关系 | 组织架构图 | [../../lark-whiteboard-cli/scenes/organization.md](../../lark-whiteboard-cli/scenes/organization.md) |
| 时间线/里程碑/版本规划 | 里程碑图 | [../../lark-whiteboard-cli/scenes/milestone.md](../../lark-whiteboard-cli/scenes/milestone.md) |
| 因果/复盘/根因分析 | 鱼骨图 | [../../lark-whiteboard-cli/scenes/fishbone.md](../../lark-whiteboard-cli/scenes/fishbone.md) |
| 方案对比/技术选型 | 对比图 | [../../lark-whiteboard-cli/scenes/comparison.md](../../lark-whiteboard-cli/scenes/comparison.md) |
| 循环/飞轮/闭环 | 飞轮图 | [../../lark-whiteboard-cli/scenes/flywheel.md](../../lark-whiteboard-cli/scenes/flywheel.md) |
| 层级占比/能力模型 | 金字塔图 | [../../lark-whiteboard-cli/scenes/pyramid.md](../../lark-whiteboard-cli/scenes/pyramid.md) |
| 模块依赖/调用关系 | 架构图 | [../../lark-whiteboard-cli/scenes/architecture.md](../../lark-whiteboard-cli/scenes/architecture.md) |
| 分类梳理/知识体系 | 思维导图 | [../../lark-whiteboard-cli/scenes/mermaid.md](../../lark-whiteboard-cli/scenes/mermaid.md) |
| 数据分布/占比 | 饼图 | [../../lark-whiteboard-cli/scenes/mermaid.md](../../lark-whiteboard-cli/scenes/mermaid.md) |

## 注意事项

- `lark-doc` 不能直接编辑已有画板内容；它负责文档层面的插入与 token 获取。
- `docs +update` 的画板能力是插入空白占位，不是画板内容编辑器。
- 只创建空白画板而不填充内容，任务视为未完成。

## 关联参考

- 画板内容生成：[`../../lark-whiteboard-cli/SKILL.md`](../../lark-whiteboard-cli/SKILL.md)
- 画板读写与更新：[`../../lark-whiteboard/SKILL.md`](../../lark-whiteboard/SKILL.md)