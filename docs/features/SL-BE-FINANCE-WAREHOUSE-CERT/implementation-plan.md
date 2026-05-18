# SL-BE-FINANCE-WAREHOUSE-CERT Implementation Plan

## Feature 信息

- Feature ID: SL-BE-FINANCE-WAREHOUSE-CERT
- 开发负责人: Codex

## 变更范围

- `internal/modules/serviceoffer/ui/handler/handler.go`
- `internal/modules/aiprocurement/application/service/finance_matcher.go`
- `docs/features/SL-BE-FINANCE-WAREHOUSE-CERT/*`

## 实施步骤

1. 让资金方我的金融产品返回授权监管仓信息。
2. 在融资匹配规则查询中按监管仓授权过滤。
3. 在向量召回和显式选择产品路径中补授权校验。

## 验证计划

- `./harness/verify.sh`

## 风险与回滚关注点

- 启用监管仓过滤后，选择监管仓但未完成资方授权时，融资产品列表会为空。
- 回滚时恢复融资匹配查询和 `ListMyFinance` 返回逻辑即可。
