# SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST Spec

## 基本信息

- Feature ID: SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST
- Feature 名称: AI采购收货地址与物流到手价匹配
- 负责人: Codex
- 目标上线环境: prod

## 背景

AI采购需要把仓储和物流能力纳入商品匹配。用户确认采购参数时应提供收货地址，该地址可以是手动地址，也可以是平台仓库地址；卖家商品如果绑定仓库，则该仓库作为发货地址参与运费估算。

## 范围

### In Scope

- `ParsedInstruction` 增加收货地址、地址来源、收货仓库标识字段。
- 商品匹配结果增加发货仓库、起运地址、推荐物流商、预估运费、到手总价、到手单价字段。
- 匹配排序从单纯价格/库存扩展为优先按到手总价排序，并保留拼单所需的候选商品数据。
- 生成订单时重新按已选商品数量估算运费，将商品金额、运费、到手价写入 AI 采购订单快照。

### Out of Scope

- 不新增收货地址簿表。
- 不改变真实交易订单表的商品金额口径，物流费用先进入 AI 采购快照。
- 不实现跨承运商实时询价或地图距离计费。

## 接口与数据影响

- 受影响路由/handler: `/api/v1/ai-procurements/confirm-instruction`, `/api/v1/ai-procurements/match-products`, `/api/v1/ai-procurements/generate-order`
- 受影响服务: `internal/modules/aiprocurement/application/service`
- 受影响 middleware: 无
- 受影响表/model: `ai_procurements.parsed_instruction`, `matched_products`, `order_draft` JSON 字段新增兼容字段
- 配置或环境变量变化: 无
- Docker / 部署影响: 无

## 权限与合规影响

- 受影响角色: 买家企业用户
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: AI采购日志新增物流到手价排序阶段；审核项不新增强制项。

## 验收标准

1. `match-products` 接收 `parsed.delivery_address` 后返回带 `freight_estimate`、`landed_total`、`logistics_service_name` 的候选商品。
2. 商品有 `warehouse_info` 时，匹配结果能展示卖家发货仓库和起运地址；无绑定仓库时不阻断匹配。
3. 多商品拼单生成订单后，AI采购快照包含所选商品、商品金额、预估运费和到手金额。
