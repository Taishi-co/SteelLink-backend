# SL-BE-PRODUCT-STRICT-ONSALE Test Cases

## 测试范围

- Feature ID: SL-BE-PRODUCT-STRICT-ONSALE
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 公开商品列表 | 存在 `status=2` 商品 | 请求 `GET /api/v1/products` | 返回已上架商品 |
| TC-02 | 显式查询已上架 | 存在 `status=2` 商品 | 请求 `GET /api/v1/products?status=2` | 只返回 `status=2` 商品 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 异常待审数据带审核时间 | 存在 `status=1` 且 `approved_at IS NOT NULL` 商品 | 请求公开商品列表 | 不返回该商品 |
| TC-EX-02 | 显式查询待审 | 存在 `status=1` 商品 | 请求 `GET /api/v1/products?status=1` | 可返回待审商品，便于排查 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 访客 | 公开商品列表 | 只能看到 `status=2` 可售商品 |
| TC-AUTH-02 | 卖家/管理查询 | 显式传状态或企业 ID | 按查询条件返回对应商品，不混用审核时间兜底 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端验证 | `./harness/verify.sh` | Go 测试和构建通过 |
| TC-OPS-02 | 发布检查 | `./harness/release.sh docs/features/SL-BE-PRODUCT-STRICT-ONSALE` | 发布门禁通过 |
