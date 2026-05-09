# COS Object Storage Uploads Test Cases

## 测试范围

- Feature ID: SL-STORAGE-COS-UPLOAD
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 上传商品主图 | 已登录，COS 环境变量已配置 | `POST /api/v1/uploads`，`purpose=product_image`，上传 jpg/png/webp | 返回 `bucket/key/url`，bucket 为公开商品桶 |
| TC-02 | 上传支付凭证 | 已登录，COS 环境变量已配置 | `POST /api/v1/uploads`，`purpose=payment_voucher`，上传图片或 PDF | 返回私有桶 URL，可提交到支付凭证接口 |
| TC-03 | 创建商品保存图片 | 已登录卖家，已获得商品主图 URL | 调用 `POST /api/v1/products`，传入 `images` 与 `quality_cert` | 商品创建成功，查询商品详情返回图片 URL |
| TC-04 | 上传企业营业执照 PDF | 已登录，COS 环境变量已配置 | `POST /api/v1/uploads`，`purpose=enterprise_license`，上传 PDF | 返回私有桶 `enterprise/licenses` URL |
| TC-05 | 上传物流商服务资质 | 已登录，COS 环境变量已配置 | `POST /api/v1/uploads`，`purpose=enterprise_service_document`，上传运输许可证/承运资质图片或 PDF | 返回私有桶 `enterprise/service-documents` URL |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未登录上传 | 无 token | 调用 `POST /api/v1/uploads` | 返回未授权 |
| TC-EX-02 | 非法上传用途 | 已登录 | `purpose=unknown` 上传文件 | 返回参数错误 |
| TC-EX-03 | 非法文件类型 | 已登录 | 商品主图上传 `.exe` | 返回文件类型错误 |
| TC-EX-04 | COS 未配置 | `STORAGE_PROVIDER=tencent_cos`，未设置 `COS_SECRET_ID` 或 `COS_SECRET_KEY` | 调用上传接口 | 返回 COS 配置错误，不创建本地 `uploads` 目录 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 游客 | 上传任意文件 | 被 `AuthRequired` 拦截 |
| TC-AUTH-02 | 已登录买家 | 上传支付/签收凭证 | 上传成功，后续订单服务继续校验订单归属 |
| TC-AUTH-03 | 已登录卖家 | 上传商品主图/审核资料 | 上传成功，后续商品创建继续校验企业身份 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 测试、构建、配置检查通过 |
| TC-OPS-02 | 环境变量示例完整 | 检查 `.env.example` | 包含 COS bucket、base URL、密钥占位符 |
| TC-OPS-03 | 本地存储显式启用 | 设置 `STORAGE_PROVIDER=local`，上传合法文件 | 写入本地目录；默认 COS 模式缺密钥时不会降级写盘 |
