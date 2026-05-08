# COS Object Storage Uploads Implementation Plan

## Feature 信息

- Feature ID: SL-STORAGE-COS-UPLOAD
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `cmd/api/main.go`
  - `config/config.go`
  - `internal/pkg/handler/upload.go`
  - `internal/pkg/storage/cos.go`
  - `internal/pkg/storage/local.go`
  - `internal/modules/product/ui/dto/product_dto.go`
  - `internal/modules/product/application/service/product_service.go`
- 文档文件:
  - `docs/features/SL-STORAGE-COS-UPLOAD/*`
- 配置文件:
  - `config/config.yaml`
  - `.env.example`
  - `go.mod`
  - `go.sum`
- 部署/脚本文件:
  - `Dockerfile`
  - 部署环境需补 COS 密钥；本地写盘仅在 `STORAGE_PROVIDER=local` 时启用。

## 实施步骤

1. 新增 `StorageConfig` 与环境变量覆盖逻辑。
2. 引入腾讯云 COS SDK，封装按 purpose 选择 bucket、key 前缀、文件类型校验的上传客户端。
3. 新增 `POST /api/v1/uploads`，复用登录态鉴权，接收 multipart 文件。
4. 扩展商品创建 DTO 与服务逻辑，保存商品图片和质保资料 URL。
5. 与前端联调支付凭证、签收凭证、商品图片上传链路。
6. 收紧 provider 选择逻辑，避免 COS 密钥缺失时静默降级到本地 `uploads` 导致容器权限错误。
7. 补齐企业营业执照和企业服务资质资料 purpose，支持物流商运输许可证、承运资质等私有桶上传。

## 验证计划

- `./harness/branch-guard.sh local`
- `./harness/verify.sh`
- `./harness/review.sh docs/features/SL-STORAGE-COS-UPLOAD`
- `./harness/release.sh docs/features/SL-STORAGE-COS-UPLOAD`
- 人工验证: 配置真实 COS 密钥后上传一张商品图、一份支付凭证、一份签收凭证、一份企业营业执照和一份物流商服务资质。

## 风险与回滚关注点

- 私有桶 URL 如果未开放读权限，前端直接点击可能返回 403；后续应补签名 URL 查看接口。
- COS 密钥必须通过环境变量注入，不能写入仓库。
- 生产环境如果没有注入 COS 密钥，接口会明确返回配置错误；不能依赖本地 fallback 保存生产资料。
- 回滚时移除上传接口和前端上传调用，恢复为传入 URL；已上传对象需按 key 前缀清理。
