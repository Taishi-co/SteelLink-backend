# SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST Test Cases

## 测试范围

- Feature ID: SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 商品匹配后匹配物流仓储 | 有上架商品、商品绑定仓库、多个已审核物流商 | 先调用 `match-products`，再用已选商品调用 `match-logistics-warehouse` | 商品匹配只返回候选商品；物流仓储匹配返回起运地址、多个 `logistics_options`、默认推荐物流商、运费和到手总价 |
| TC-02 | 多商品拼单生成订单快照 | 已选 A/B 两个商品且数量合计等于采购量 | 调用 `generate-order` | 真实订单按卖家拆单，AI采购 `order_draft.order_detail` 包含商品金额和预估运费 |
| TC-03 | 平台仓库作为收货地址 | 前端传入 `delivery_address_source=buyer_warehouse` | 调用物流仓储匹配接口 | 匹配排序使用该仓库地址作为目的地 |
| TC-04 | 用户选择非默认物流 | `match-logistics-warehouse` 返回多个可选方案 | 前端把用户选择的 `logistics_service_id` 传给 `generate-order` | 生成订单前重算时保留用户选择的物流方案 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未传收货地址 | 只传原有解析参数 | 调用 `match-logistics-warehouse` | 接口兼容旧请求，返回商品但运费为空或为待估算 |
| TC-EX-02 | 无适合物流商 | 商品存在但没有覆盖起运地和目的地的物流服务 | 调用 `match-logistics-warehouse` | 返回 `logistics_match_status=no_match` 和空 `logistics_options`，不伪造默认物流商 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录用户 | 调用只读匹配接口 | 保持现有接口权限策略 |
| TC-AUTH-02 | 非企业用户 | 调用 `generate-order` | 返回企业账号校验错误 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端构建验证 | `./harness/verify.sh` | Go 测试和构建通过 |
