# Test Cases

## 测试范围

- Feature ID: SL-BE-BIDDING-PROCUREMENT-API
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 查询竞价统计 | 后端服务启动 | 请求 `GET /api/v1/biddings/stats` | 返回进行中、成功、成交金额、参与企业统计 |
| TC-02 | 查询竞价列表 | 后端服务启动 | 请求 `GET /api/v1/biddings?page=1&page_size=20` | 返回竞价活动分页列表 |
| TC-03 | 查询竞价详情 | 存在竞价活动 | 请求 `GET /api/v1/biddings/:id` | 返回指定竞价活动 |
| TC-04 | 查询集采统计 | 后端服务启动 | 请求 `GET /api/v1/procurements/stats` | 返回活动数、筹备数、成交金额、参与企业等统计 |
| TC-05 | 查询集采列表 | 后端服务启动 | 请求 `GET /api/v1/procurements?page=1&page_size=12` | 返回集采活动分页列表 |
| TC-06 | 查询集采详情 | 存在集采活动 | 请求 `GET /api/v1/procurements/:id` | 返回活动详情和活动日志 |
| TC-07 | 竞价交易闭环 | 两个不同企业均已认证且已登录 | 发起方企业创建竞价，另一家企业报价，截止后发起方定标，再生成订单 | 状态依次为 `bidding`、`pending_award`、`awarded`、`order_created`，订单返回 `order_id/order_no` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 查询不存在竞价 | 后端服务启动 | 请求不存在的 `GET /api/v1/biddings/:id` | 返回 404 |
| TC-EX-02 | 查询不存在集采 | 后端服务启动 | 请求不存在的 `GET /api/v1/procurements/:id` | 返回 404 |
| TC-EX-03 | 分页参数缺省 | 后端服务启动 | 不传 `page/page_size` 请求列表 | 默认按第一页、10 条返回 |
| TC-EX-04 | 截止前定标 | 存在 `bidding` 状态竞价且已有报价 | 发起方请求 `POST /api/v1/biddings/:id/award` | 返回业务错误，提示竞价截止并进入待定标后才能定标 |
| TC-EX-05 | 参与自身竞价 | 企业已发起竞价 | 同一企业请求 `POST /api/v1/biddings/:id/quotes` | 返回业务错误，提示发起方不能参与自身竞价报价 |
| TC-EX-06 | 高于当前最低价报价 | 已存在更低报价 | 报价方提交高于当前最低价的报价 | 返回业务错误，提示报价不能高于当前最低价 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 访问竞价/集采公开接口 | 可正常读取大厅公开数据 |
| TC-AUTH-02 | 未登录用户 | 访问竞价公开列表和详情 | 不返回完整报价明细，不展示未成交供应商身份 |
| TC-AUTH-03 | 已登录发起方企业 | 访问自己发起的竞价工作台数据 | 返回完整报价明细和供应商身份 |
| TC-AUTH-04 | 已登录报价方企业 | 访问竞价工作台数据 | 只返回自己的报价和中标结果，不暴露其他未中标供应商身份 |

## 本地验收脚本

已提供脚本：`scripts/verify-bidding-flow.sh`

运行方式：

```bash
BUYER_TOKEN=... SELLER_TOKEN=... bash scripts/verify-bidding-flow.sh
```

脚本会验证：

- 发起方企业创建竞价。
- 另一家企业提交报价。
- 截止前定标被拒绝。
- 截止后查询触发 `pending_award`。
- 发起方定标进入 `awarded`。
- 发起方生成订单进入 `order_created`。

## 回归关注点

- 商品、订单、用户现有接口不受影响。
- AutoMigrate 不应破坏已有 `products`、`orders` 等交易表。
- 竞价交易不依赖企业 `enterprise_role` 单选身份；同一个已认证企业账号可发起采购，也可在其他企业发起的竞价中报价。
- 公开大厅不能因登录态存在而泄露报价明细；只有 `scope=workspace` 的企业工作台请求可返回角色视角数据。
