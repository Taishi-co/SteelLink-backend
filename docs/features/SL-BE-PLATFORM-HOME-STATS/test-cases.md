# Platform Home Stats Test Cases

## 测试范围

- Feature ID: SL-BE-PLATFORM-HOME-STATS
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 查询首页统计 | 后端服务启动且数据库可用 | 请求 `GET /api/v1/platform/stats` | 返回 `merchant_count`, `sku_count`, `order_count`, `transaction_amount` |
| TC-02 | 空数据库统计 | 相关表无业务数据 | 请求 `GET /api/v1/platform/stats` | 四项统计返回 0，不报错 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 数据库不可用 | MySQL 查询失败 | 请求 `GET /api/v1/platform/stats` | 返回统一内部错误响应 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 游客 | 未携带 token 访问统计接口 | 可正常读取聚合统计 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 构建通过，配置解析无阻断 |
