# Release Checklist Template

## 发布基本信息

- Feature ID:
- 发布负责人:
- 目标环境:

## 发布前检查

- [ ] `./harness/release.sh docs/features/<feature-id>` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] 如环境安装了 Docker Compose，`docker compose config` 已通过
- [ ] 已确认配置与环境变量变更
- [ ] 已确认数据库影响
- [ ] 已确认 smoke test 范围

## 回滚信息

- 回滚目标版本:
- 回滚步骤:
- 回滚后验证:
