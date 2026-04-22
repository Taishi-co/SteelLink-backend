# Feature Spec

## 基本信息

- Feature ID: `SL-BE-WIN-REMOTE-BOOTSTRAP`
- Feature 名称: 主业务后端 Windows 远端依赖启动脚本
- 负责人: 二号开发者
- 目标上线环境: 本地开发环境

## 背景

主业务后端在 Windows 本地联调时，需要连接服务器上的 MySQL 和 Redis。
当前仓库缺少统一的 Windows 启动脚本，导致本地调试需要重复手工设置环境变量。

本次改动仅补齐 Windows 本地联调脚本，不触碰线上配置。

## 范围

### In Scope

- 为主业务后端新增 Windows 启动脚本入口，封装远端 MySQL / Redis 联调环境变量。
- 在 `.env.example` 中补充远端依赖开发说明。

### Out of Scope

- 数据库表结构修改。
- 服务器配置改动。
- 业务接口逻辑变更。

## 页面与交互影响

- 受影响页面: 无
- 受影响路由: 无
- 受影响布局或导航: 无
- 受影响组件: 无
- 受影响交互状态: 无

## 接口与数据影响

- 受影响接口: 无直接接口变更
- 请求参数或响应结构变化: 无
- 受影响服务层: 启动入口
- 受影响状态管理: 无
- 受影响后端 handler/service/repository: 无
- 受影响数据库或缓存: 仅连接方式说明变化
- 配置或环境变量变化:
  - `DATABASE_HOST`
  - `DATABASE_PORT`
  - `DATABASE_USERNAME`
  - `DATABASE_PASSWORD`
  - `DATABASE_NAME`
  - `REDIS_HOST`
  - `REDIS_PORT`
  - `REDIS_PASSWORD`
- 构建或部署影响:
  - 仅新增 Windows 本地启动脚本入口

## 权限与合规影响

- 受影响角色: 无
- 新增/调整权限: 无
- 是否影响审核、风控、日志留痕: 否

## 验收标准

1. Windows 下可通过脚本一键设置远端 MySQL / Redis 环境变量并启动后端。
2. 脚本不修改服务器配置，也不写入线上环境。
3. 后端 `go build ./...` 通过。
