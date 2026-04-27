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

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 查询不存在竞价 | 后端服务启动 | 请求不存在的 `GET /api/v1/biddings/:id` | 返回 404 |
| TC-EX-02 | 查询不存在集采 | 后端服务启动 | 请求不存在的 `GET /api/v1/procurements/:id` | 返回 404 |
| TC-EX-03 | 分页参数缺省 | 后端服务启动 | 不传 `page/page_size` 请求列表 | 默认按第一页、10 条返回 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 访问竞价/集采公开接口 | 可正常读取大厅公开数据 |
| TC-AUTH-02 | 已登录买家 | 访问竞价/集采公开接口 | 可正常读取同一份公开数据 |

## 回归关注点

- 商品、订单、用户现有接口不受影响。
- AutoMigrate 不应破坏已有 `products`、`orders` 等交易表。
