# LTerm 电影院订票系统 · 文档总览

> 项目代号：LTerm | 技术栈：Go(Gin) + PostgreSQL + React(TS) + ECharts + Docker Compose
> 定位编号规则：`B0000M00000P0000M00F00`，可沿编号定位任意层级文档/功能段落

## 文档目录

| 层级 | 文档 | 说明 | 状态 |
| --- | --- | --- | --- |
| L1 战略 | [BRD/B0000-电影院订票系统.md](BRD/B0000-电影院订票系统.md) | 为什么做、商业价值、投入产出 | 草稿 v0.1 |
| L2 战术 | [MRD/M0000-电影院订票系统.md](MRD/M0000-电影院订票系统.md) | 用户/市场/竞品/产品策略 | 草稿 v0.1 |
| L3 需求 | [PRD/P0000-电影院订票系统.md](PRD/P0000-电影院订票系统.md) | 开发就绪需求：功能、规则、验收、状态机 | 优化稿 v0.2 |
| L4 详设 | [FSD/F0000-订单支付与退款.md](FSD/F0000-订单支付与退款.md) | 支付/退款/积分联动详设（本次重点） | 详设 v0.2 |
| 数据 | [database/数据库表.md](database/数据库表.md) | 全量表结构、ER 图、幂等与并发设计（无钱包） | 草稿 v0.2（待你评审） |
| 状态 | [state-machine/状态机.md](state-machine/状态机.md) | 订单/支付/退款/选座/优惠券状态机全集 | 草稿 v0.2 |
| 排期 | [schedule/排期.md](schedule/排期.md) | 里程碑/迭代排期风格与明细 | 草稿 v0.1 |
| 学习 | [guide/业务建模与贫血DDD.md](guide/业务建模与贫血DDD.md) | 业务建模四步法 + 贫血 DDD 用例编排 + 练习 | v0.1 |
| 架构 | [guide/项目架构-贫血DDD.md](guide/项目架构-贫血DDD.md) | Go 项目目录结构、分层职责、事务所有权 | v0.1 |
| 参考 | [guide/如何参考开源项目.md](guide/如何参考开源项目.md) | 主流架构对比、推荐项目、五步阅读法、练习任务 | v0.1 |
| 蓝图 | [guide/LTerm后端结构蓝图.md](guide/LTerm后端结构蓝图.md) | 对齐 potentia/backend：core/domain + core/port + infra + app/server | v0.2 |

## 协作约定

1. 逐步推进：先冻结「数据模型 + 状态机」，再细化功能与 FSD，最后进入开发。
2. 数据库表文档由你评审定稿，我负责后续优化（索引、约束、幂等、对账）。
3. 金额一律以「分」为单位存储；状态流转必须走服务层状态机，禁止直接 UPDATE；**不做钱包/虚拟货币**。
4. 所有图表使用 Mermaid，规则见 `rules/`。

## 规则与草稿

- 文档模板规则：`rules/prd-format.md`、`rules/brd-format.md`、`rules/mrd-format.md`
- 角色设定：`rules/role.md`（产品架构师 Tyrion）
- 视觉规范：`rules/design-system.md`（黑金高级简约）
- 需求原始草稿：`PRD/需求原始草稿.md`
