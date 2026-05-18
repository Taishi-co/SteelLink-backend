# SL-BE-FINANCE-WAREHOUSE-CERT Spec

## 基本信息

- Feature ID: SL-BE-FINANCE-WAREHOUSE-CERT
- Feature 名称: 监管仓认证与金融产品匹配后端闭环
- 负责人: Codex
- 目标上线环境: staging

## 背景

后端已有资金方授权监管仓接口和 `finance_authorized_warehouses` 表，但 AI 采购融资匹配仍按金额和产品规则匹配全部金融产品，没有按用户选择的监管仓限制候选产品。

## 范围

### In Scope

- 资金产品列表返回已授权监管仓 ID。
- AI 采购融资匹配按 `ParsedInstruction.delivery_warehouse_id` 过滤对应金融产品。
- 用户显式选择金融产品时校验其是否授权所选监管仓。

### Out of Scope

- 不新增数据库表。
- 不调整平台审核流程。

## 接口与数据影响

- 受影响接口:
  - `GET /api/v1/finance-services/mine`
  - `POST /api/v1/ai-procurements/:id/match-finance`
- 受影响表/model:
  - `finance_services`
  - `finance_authorized_warehouses`
- 权限影响: 复用资金方 `finance` 服务能力。

## 验收标准

1. 金融产品详情包含 `authorized_warehouse_ids`。
2. AI 采购选择监管仓后，只返回该仓授权的金融产品。
3. 未授权的显式选择产品不会被追加到融资方案。
