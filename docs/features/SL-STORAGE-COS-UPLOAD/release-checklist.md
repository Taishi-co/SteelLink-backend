# COS Object Storage Uploads Release Checklist

## 发布基本信息

- Feature ID: SL-STORAGE-COS-UPLOAD
- 发布负责人: Codex
- 目标环境: staging, prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-STORAGE-COS-UPLOAD` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认生产环境配置 `COS_SECRET_ID`
- [ ] 已确认生产环境配置 `COS_SECRET_KEY`
- [ ] 已确认三个 bucket 与 `ap-beijing` 地域匹配
- [ ] 已确认 smoke test 范围: 上传商品图、上传支付凭证、上传货权/监管凭证、上传签收凭证、上传企业营业执照、上传物流商运输许可证和承运资质

## 回滚信息

- 回滚目标版本: 合并前主干版本
- 回滚步骤: 回滚后端上传接口和前端调用，重新部署后端和前端
- 回滚后验证: 商品列表、订单详情、支付/货权监管/签收流程恢复到回滚版本行为
