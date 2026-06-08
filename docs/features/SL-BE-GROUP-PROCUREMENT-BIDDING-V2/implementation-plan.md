# Implementation Plan

## Feature 信息

- Feature ID: `SL-BE-GROUP-PROCUREMENT-BIDDING-V2`
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/market/domain/entity/*`
  - `internal/modules/market/ui/dto/*`
  - `internal/modules/market/application/service/*`
  - `internal/modules/market/ui/handler/*`
  - `internal/modules/market/infrastructure/router/*`
  - `internal/modules/market/infrastructure/persistence/*`
- 文档文件:
  - `docs/features/SL-BE-GROUP-PROCUREMENT-BIDDING-V2/*`
- 配置文件:
  - 无
- 部署/脚本文件:
  - 无

## 实施步骤

1. 新增项目、报名、报价、定标实体和迁移，补齐 DTO 与状态枚举。
2. 实现项目创建、供应商报名/报价、管理端审核/定标/成单服务与路由，并兼容旧集采活动读接口。
3. 增加日志留痕、权限校验、截止时间校验和基于中标报价的订单生成逻辑。

## 验证计划

- `./harness/verify.sh`
- 如涉及镜像或容器：`docker compose config`
- 需要的人工验证:
  - 管理端创建项目并邀请供应商
  - 卖家报名与报价
  - 管理端定标与成单

## 风险与回滚关注点

- 旧 `procurement_activities` 与新项目模型并存，查询兼容要避免相互污染。
- 单供应商中标成单依赖报价绑定真实商品，若数据映射不完整会阻断成单。
- 回滚时应仅下线路由与服务逻辑，保留新增表不影响旧功能读写。
