# SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE Test Cases

## 测试范围

- Feature ID: SL-BE-SERVICE-MARKET-FULFILLMENT-INTAKE
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 提交仓储托管需求 | 已登录企业用户，存在已审核仓库 | `POST /warehouse-storage-requests` | 返回 `movement_type=inbound`、`status=pending` 的记录 |
| TC-02 | 查询我的托管需求 | 企业已提交托管需求 | `GET /warehouse-storage-requests` | 只返回当前企业发起的需求 |
| TC-03 | 仓储方处理作业 | 仓储方有对应仓库服务 | `GET /warehouse-movements?role=warehouse` 后更新状态 | 能看到并处理仓库名下作业 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 仓库未审核 | 仓库状态非已通过 | 提交托管需求 | 返回仓储资源不存在或未审核通过 |
| TC-EX-02 | 未登录 | 无 token | 提交托管需求 | 鉴权失败 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 企业用户无服务能力 | 调用托管接口 | 被服务能力 middleware 拦截 |
| TC-AUTH-02 | 仓储方 | 查看 `role=warehouse` 出入库记录 | 只看当前企业仓库作业 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端验证 | `./harness/verify.sh` | Go 测试、构建和 compose config 通过 |
