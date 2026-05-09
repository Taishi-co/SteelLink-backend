# Redis Cache Expansion Test Cases

## 测试范围

- Feature ID: SL-BE-REDIS-CACHE-EXPANSION
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 商品公开列表缓存 | Redis 可用，存在上架商品 | 连续调用商品列表接口 | 第二次可从 Redis 返回同等响应 |
| TC-02 | 服务资源列表缓存 | Redis 可用，存在 approved 仓储/物流/资金服务 | 连续调用服务列表接口 | 第二次可从 Redis 返回同等响应 |
| TC-03 | 企业服务权限缓存 | 企业已开通指定服务 | 调用受 `RequireEnterpriseService` 保护的接口 | 权限判断成功并写入短缓存 |
| TC-04 | AI 采购匹配缓存 | Redis 可用，同一 parsed instruction 重复匹配 | 连续调用匹配接口 | 第二次返回同样匹配结果且不改变订单主数据 |
| TC-05 | refresh token 轮换 | 用户持有有效 refresh token | 调用刷新接口后再次使用旧 token | 第二次返回 token invalid |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | Redis 缓存 miss | Redis 中无对应 key | 调用商品/服务/市场接口 | 回源 MySQL 并正常返回 |
| TC-EX-02 | Redis 不可用 | Redis 连接失败 | 调用商品/服务/市场接口 | 除短信和 AI 配置强依赖链路外，不影响主流程 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未开通服务企业 | 调用受服务能力保护接口 | 返回无权访问 |
| TC-AUTH-02 | 权限审核更新后企业 | Admin 审核后再次调用接口 | 对应企业权限缓存失效，按最新权限判断 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 构建通过，配置解析无阻断 |
| TC-OPS-02 | Redis 配置复用 | 检查 `config/config.yaml` 的 `redis` 段 | 无新增必填配置 |
