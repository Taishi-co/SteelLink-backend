# SL-BE-FINANCE-WAREHOUSE-CERT Test Cases

## 测试范围

- Feature ID: SL-BE-FINANCE-WAREHOUSE-CERT
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 监管仓过滤融资产品 | 金融产品 A 已授权仓库 X，产品 B 未授权 | AI 采购选择仓库 X 后匹配融资 | 仅返回 A |
| TC-02 | 金融产品返回授权仓 | 资金方已有授权记录 | 请求我的金融产品 | 返回 `authorized_warehouse_ids` |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 显式选择未授权产品 | 选择监管仓 X，但产品 B 未授权 | 请求融资匹配并传 B | B 不进入结果 |

## 接口与后端验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-BE-01 | 后端测试 | `./harness/verify.sh` | 测试和构建通过 |
