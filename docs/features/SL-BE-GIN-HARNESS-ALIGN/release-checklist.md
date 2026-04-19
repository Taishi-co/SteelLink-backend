# Release Checklist

## 发布基本信息

- Feature ID: `SL-BE-GIN-HARNESS-ALIGN`
- 发布负责人: 待定
- 目标环境: staging

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-GIN-HARNESS-ALIGN` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变化仅限工程流程变量
- [ ] 已确认数据库影响为无
- [ ] 已确认 smoke test 范围

## 回滚信息

- 回滚目标版本: 迁移前的 harness / docs / Dockerfile 版本
- 回滚步骤: 回退 `harness/`、`docs/`、`Dockerfile`、`docker-compose.yml`
- 回滚后验证: `./harness/verify.sh` 与文档内容恢复到目标版本
