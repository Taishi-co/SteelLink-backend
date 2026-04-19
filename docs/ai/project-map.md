# Backend Project Map

## 项目定位

`SteelLink-backend` 当前是一个基于 Go 的后端重写仓库，主入口在 `cmd/api/main.go`，Web 框架使用 Gin，配置通过 Viper 加载，日志使用 Zap，数据库访问层已接入 GORM(MySQL) 与 Redis。

## 技术栈

- 语言：Go 1.25
- Web 框架：Gin
- ORM：GORM
- 数据库：MySQL
- 缓存：Redis
- 认证：JWT
- 配置：Viper
- 日志：Zap

## 目录地图

- `cmd/api/main.go`
  应用启动入口，负责加载配置、初始化 logger / MySQL / Redis、注册 Gin 路由并处理优雅停机。
- `config/`
  配置结构与示例配置，`config.go` 定义全局配置对象与环境变量映射规则。
- `internal/pkg/database/`
  MySQL 与 Redis 初始化逻辑，连接池参数和连接错误会直接影响服务启动。
- `internal/pkg/handler/`
  当前放置 HTTP handler，已包含健康检查接口。
- `internal/pkg/middleware/`
  Gin 中间件，包括鉴权、日志、恢复、跨域。
- `pkg/`
  公共能力层，当前包含 JWT、错误码、日志封装、统一响应结构。
- `harness/`
  项目约束脚本入口，负责 bootstrap、分支检查、验证、review、release、deploy。
- `docs/ai/`
  agent 协作规则与工作流文档。
- `docs/features/`
  feature 设计、测试、实施、发布文档；`_template/` 为模板，非模板目录才代表真实需求。
- `docker-compose.yml`
  本地依赖与容器编排定义，当前包含 `api`、`mysql`、`redis`。
- `Dockerfile`
  服务镜像构建入口，release / deploy 统一以此文件为准。

## 关键模块

- 配置加载：`config/config.go`
- HTTP 启动与路由：`cmd/api/main.go`
- 数据库连接：`internal/pkg/database/mysql.go`
- Redis 连接：`internal/pkg/database/redis.go`
- 鉴权与 claims：`pkg/auth/jwt.go`、`internal/pkg/middleware/auth.go`
- 统一日志：`pkg/logger/logger.go`、`internal/pkg/middleware/logger.go`
- 统一响应与错误码：`pkg/response/response.go`、`pkg/ecode/codes.go`

## 高风险改动区域

- `config/` 与 `config/*.yaml`
  配置字段名、环境变量映射和默认值一旦变更，会直接影响启动、部署和容器配置。
- `internal/pkg/database/`
  MySQL / Redis 连接参数、DSN、连接池策略和启动时失败逻辑都在这里。
- `pkg/auth/` 与 `internal/pkg/middleware/auth.go`
  涉及 token 签发、解析和接口鉴权路径，任何调整都要同步验证权限矩阵。
- `Dockerfile` 与 `docker-compose.yml`
  会影响本地联调、镜像构建和部署入口，变更前后都需要明确 smoke test。
- `harness/`
  是团队统一流程入口，修改时必须同步更新 `AGENTS.md` 与 `docs/ai/*.md`。

## 变更前检查

在以下场景必须先补文档再开发：

- 新增或调整路由、handler、middleware
- 新增或修改数据库表、GORM model、Redis key 规则
- 调整 JWT、权限、审批、风控、审计相关逻辑
- 修改配置结构、环境变量、Dockerfile、docker-compose、部署流程
- 迁移 Git 工作流、仓库 remote、默认分支或 PR 规则
