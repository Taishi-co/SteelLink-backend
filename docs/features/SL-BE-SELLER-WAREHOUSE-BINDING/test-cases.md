# Seller Warehouse Binding Backend Test Cases

## 测试范围

- Feature ID: SL-BE-SELLER-WAREHOUSE-BINDING
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 销售企业创建仓库 | 登录企业具备 `sales` 能力 | 调用 `POST /warehouse-services` | 创建成功，返回状态为待审核 |
| TC-02 | 销售企业查询自有仓库 | 登录企业具备 `sales` 能力且已有仓库 | 调用 `GET /warehouse-services/mine` | 返回本企业仓库资源列表 |
| TC-03 | 销售企业编辑仓库 | 登录企业具备 `sales` 能力且仓库属于本企业 | 调用 `PUT /warehouse-services/:id` | 更新成功并重新进入待审核 |
| TC-04 | 审核通过仓库绑定商品 | 仓库资源已审核通过 | 商品创建请求携带 `warehouse_service_id` | 商品保存仓库快照和仓储资源 ID |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 绑定未审核仓库 | 仓库资源为待审核 | 商品创建请求携带该 `warehouse_service_id` | 返回“绑定的仓储资源尚未审核通过” |
| TC-EX-02 | 非所属企业编辑仓库 | 仓库属于其他企业 | 调用 `PUT /warehouse-services/:id` | 返回不存在或无权修改 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 调用仓库写接口 | 返回未登录 |
| TC-AUTH-02 | 无 `sales`/`warehouse` 能力企业 | 调用仓库写接口 | 返回无权访问 |
| TC-AUTH-03 | 仓储服务商企业 | 调用原仓库写接口 | 保持可用 |

## 回归关注点

- 物流、金融资源写接口仍分别要求 `logistics`、`finance` 服务商能力。
- 公共仓库资源列表和详情不受影响。
