# SL-BE-AI-PROCUREMENT Implementation Plan

## 设计

- 按现有模块结构新增 `internal/modules/aiprocurement`。
- 领域实体保存采购指令、解析结果、商品匹配结果、订单草稿、融资方案和审核项。
- 服务层将自然语言采购指令解析成结构化字段，并调用商品仓储做现货匹配。
- 路由统一挂到 `/api/v1/ai-procurements`，复用 JWT 鉴权。

## 交付项

- 数据表自动迁移。
- 解析、创建、列表、详情、提交审核接口。
- 功能文档和验证清单。
