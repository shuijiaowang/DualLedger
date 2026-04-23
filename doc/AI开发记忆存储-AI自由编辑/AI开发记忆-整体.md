# AI开发记忆（DualLedger）

## 1. 项目定位
- 项目名称：双记账系统（DualLedger）。
- 核心目标：同时支持两种记账视角，帮助用户既看见"钱何时流动"，也看见"价值何时被消耗"。
  - 现金流视图：记录资金实际收支发生时间。
  - 权责视图：记录消费价值的实际消耗时间，支持分摊/摊销。

## 2. 技术栈理解
- 后端：Go + Gin + MySQL（本地 8.0，服务器 5.7）（目录 `service/`）。
- PC Web 前端：Vue3 + Vite + Naive UI（目录 `web/`）。
- App/小程序前端：uni-app（目录 `web_uniapp/`，阶段优先级靠后）。
- 认证：已有 JWT、注册/登录、示例鉴权接口等基础能力可扩展。

## 3. 当前需求优先级
- 第一优先：后端完善到约 90%（数据库设计、核心接口、业务规则稳定）。
- 第二优先：Web 端测试实用优先，完成可测试的基础页面与流程，目标约 60%（"能看能用"为主，样式可简化）。
- 第三优先：uni-app 在后端与 Web 基本稳定后再迁移/复用实现。

## 4. 数据库设计版本状态（重要）

**当前版本：v2 精简版（2026-04-23 开发者审查后重写）**

- 权威文档：`doc/AI开发记忆存储-AI自由编辑/数据库设计.md`
- 通俗对照：`doc/开发者我自己的文档-仅可读/数据库设计-通俗版.md`
- 场景演示：`doc/AI开发记忆存储-AI自由编辑/数据流示例.md`

### 4.1 v2 相对 v1 的核心精简决定

基于开发者（`doc/开发者我自己的文档-仅可读/个人理解.md`）的意见，v2 做了以下**不可逆**简化，后续开发必须遵循：

1. **砍掉 cash_entry 表**：transaction 自带 `account_id / to_account_id / direction`，单账户交易不再派生 cash_entry；转账通过 `to_account_id` 表达双侧。
2. **accrual_entry 只存真实事件**：**去掉 source=AUTO**。规则性摊销（FIXED_PERIOD / DYNAMIC_BY_DAY）不写行，由服务端查询时动态计算。accrual_entry 只承载 `MANUAL / ADJUST / END_SETTLE`。
3. **砍掉 transaction_tag / accrual_entry_tag 中间表**：标签改为业务行 `ext_json.tags`（或同语义 JSON 字段）直接存储。
4. **去掉虚拟账户概念**：`account.is_virtual / type` 字段删除；借出/押金/代付用 transaction 上的 `counterparty` 文本字段 + 对应 `type` + `direction` 表达。
5. **简化 ResourceStatus**：去掉 PAUSED，保留 ACTIVE / ENDED / RETURNED / DISCARDED。
6. **简化 amortize_rule.type**：从 5 种砍到 3 种——`FIXED_PERIOD / BY_COUNT / DYNAMIC_BY_DAY`；SCHEDULED、独立的 MANUAL 合并入 BY_COUNT 的打卡行为。
7. **category 最终定稿（兼容未来多用户）**：MVP 仅系统分类；可在分类表预留 `owner_user_id`（MVP 为 NULL）并用 `source=system/user` 区分来源，后续放开用户自定义时迁移成本可控（不采用预留 id 段）。
8. **tag 最终定稿**：MVP 不建独立 tag 表，不做中间表；标签统一落在业务行 `ext_json.tags`（或同语义 JSON 字段），允许用户自定义语义标签（如 `egg_buy`、`egg_eat`）。
9. **标签统计口径定稿**：先按标签过滤/匹配，再按结构化字段（`qty/unit/amount`）统计；不把数量语义写入标签后缀。
10. **删除颜色等非必要字段**：`category.color / tag.color / account.type / account.is_virtual` 全部删除。

### 4.2 最终表清单（MVP）

5 张业务表：`user / account / resource / transaction / accrual_entry`。

Category 与 Tag 走代码常量（或只读 seed 表）。

### 4.3 关键业务规则记忆锚点

- **动态计算权责视图**（核心）：查询某区间的权责条目时，服务端返回两部分合并结果：
  1. 对所有 `status=ACTIVE/ENDED` 且 `amortize_rule.type ∈ {FIXED_PERIOD, DYNAMIC_BY_DAY}` 的 resource，按规则动态生成只读虚拟行。
  2. 区间内的 `accrual_entry` 真实事件行（正负都可能）。
  两者按分类/标签聚合相加，在终态时精确闭合 = `resource.total_cost`。
- **标签继承**：写 accrual_entry 时若有 transaction/resource 上下文，默认把 `tags` JSON 数组拷贝过去；用户可编辑覆盖，不回写源。
- **账户余额**：由应用层每次写入 transaction 时同步维护；提供"按 transaction 重算"的后台接口。
- **转账不走权责**：TRANSFER / LOAN / DEPOSIT / REFUND 默认在权责视图隐藏，前端通过"全部视图"开关才显示（仅展示不计入合计）。
- **工资与请假**：工资当作 `INCOME + resource + FIXED_PERIOD`；请假/迟到写 `type=ADJUST` 的 transaction + `source=ADJUST` 的 accrual_entry 负值。不需要每日定时任务。

### 4.4 后端代码布局（v2）

新增/保留：
```
service/model/
  common.go            // gorm Base、Money 类型别名
  user.go              // 已有
  account.go           // 新增
  resource.go          // 新增
  transaction.go       // 新增
  accrual_entry.go     // 新增
  enum.go              // 所有枚举
  presets_category.go  // 分类代码常量
  presets_tag.go       // 标签建议词常量

service/service/
  onboarding_service.go     // 新增：新用户建主账户
  ledger_service.go         // 新增：事务内写 transaction + 同步账户余额 + 衍生 accrual_entry
  accrual_view_service.go   // 新增：动态计算 + 合并真实事件
```

删除/不再新建：
- **不要** `cash_entry.go / category.go / tag.go / transaction_tag.go / accrual_entry_tag.go`（v2 去掉）。
- `service/model/example.go` 可保留或删除，不进 AutoMigrate。

AutoMigrate 列表：`User, Account, Resource, Transaction, AccrualEntry`。

## 5. 任务计划（AI 协作版本）
1. 按 v2 设计落地 5 个 model + enum + 预设常量。
2. 先实现核心 API：认证（已有）、账户 CRUD、transaction 录入（EXPENSE/INCOME/TRANSFER 三种先通）、现金流列表、账户余额重算接口。
3. 再实现 resource + 动态计算权责视图接口、BY_COUNT 打卡接口、REFUND/ADJUST 特殊类型。
4. 通过 Web 前端快速联调：登录、记一笔、列表（双视图切换）、资源管理、打卡。
5. 补齐异常处理与权限校验，形成可回归的最小闭环。
6. 核心流程稳定后再推 uni-app。

## 6. 关键约束与协作原则
- **先"宏观可控"再"快速编码"**：每轮开发前先说明思路、阶段目标、影响范围。
- **业务正确性优先于界面精细度**：先保证数据与规则正确，再优化体验。
- **接口与配置一致化**：端口、鉴权、错误码、响应结构统一。
- **新功能围绕双视图模型**：避免偏离"现金流 + 权责"核心价值。
- **每轮迭代可运行可验证**：避免一次性大改。
- **不变式驱动开发**：关键业务不变式（§见 数据库设计.md 的第 15 节相关）每加一个模块都要写 1~2 条 e2e 测试兜底。

## 7. 下一步建议（可立即执行）
- 在 `service/model/` 下按 §4.4 布局创建 5 个 model 文件 + enum.go + 两个 presets 文件。
- 在 `service/db/db.go` 里把 AutoMigrate 列表更新为 v2 的 5 个模型。
- 先打通 `POST /transactions`（支持 EXPENSE/INCOME/TRANSFER 基础三种）+ `GET /transactions`（现金流列表）+ `GET /accounts` 这三个核心接口。
- Web 端先做 3 个最小页面：登录页、记账页、记录页，端到端走通。
