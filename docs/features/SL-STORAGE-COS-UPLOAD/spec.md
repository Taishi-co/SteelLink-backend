# COS Object Storage Uploads

## 基本信息

- Feature ID: SL-STORAGE-COS-UPLOAD
- Feature 名称: 腾讯云 COS 文件上传接入
- 负责人: Codex
- 目标上线环境: staging, prod

## 背景

当前商品主图、支付凭证、签收凭证等文件类资料仍依赖前端传入 URL 或占位数据，无法形成真实上传链路。平台已经具备三个 COS 桶，需要后端提供统一上传入口，并按业务用途写入公开或私有桶。

## 范围

### In Scope

- 新增鉴权后的 multipart 文件上传接口。
- 按上传用途选择 COS 桶和对象 key 前缀。
- 支持商品主图、商品审核资料、支付凭证、签收凭证、企业营业执照和企业服务资质资料。
- 商品创建接口保存商品主图 URL 和质保/审核资料 URL。
- 增加 COS 配置项与环境变量示例。

### Out of Scope

- 私有桶文件的临时签名查看接口。
- 文件病毒扫描、水印处理、图片压缩。
- 管理端独立上传或预览体验改造。

## 接口与数据影响

- 受影响路由/handler: `POST /api/v1/uploads`, `POST /api/v1/products`
- 受影响服务: `internal/pkg/storage`, `internal/modules/product/application/service`, `internal/modules/user/application/service`
- 受影响 middleware: 上传接口使用 `AuthRequired`
- 受影响表/model: `products.images`, `products.quality_cert`
- 配置或环境变量变化: 新增 `storage` 配置与 `COS_*` / `STORAGE_*` 环境变量
- Docker / 部署影响: 部署环境必须注入 `COS_SECRET_ID` 与 `COS_SECRET_KEY`；只有显式配置 `STORAGE_PROVIDER=local` 时才允许写入本地 `uploads` 目录。

## 权限与合规影响

- 受影响角色: 已登录买家、卖家。
- 新增/调整权限: 未新增细粒度权限，上传接口复用登录态鉴权。
- 是否影响审核、风控、日志留痕: 商品主图影响商品审核资料完整性；支付/签收凭证影响订单履约证据留存；企业营业执照和服务资质影响企业/服务商审核资料完整性。

## 验收标准

1. 已登录用户可上传合法类型文件并获得 COS URL。
2. 商品创建请求必须包含至少一张商品主图，后端保存真实 URL。
3. 支付凭证和签收凭证由前端上传后提交 URL，订单数据不再依赖手填占位 URL。
4. `STORAGE_PROVIDER=tencent_cos` 缺少 COS 密钥或上传非法类型文件时返回明确错误，不能静默降级为本地 `uploads` 写盘。
5. 企业创建/认证中的营业执照、运输许可证、承运资质等资料可用 `enterprise_license` 或 `enterprise_service_document` 上传，支持图片和 PDF。
