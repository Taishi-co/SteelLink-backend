# Dockerfile Backend Deployment Script Release Checklist

## 发布基本信息

- Feature ID: SL-BE-DOCKERFILE-DEPLOY
- 发布负责人: Codex
- 目标环境: prod

## 发布前检查

- [ ] `./harness/release.sh docs/features/SL-BE-DOCKERFILE-DEPLOY` 已通过
- [ ] `go test ./...` 已通过
- [ ] `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 已通过
- [ ] `bash -n deploy_backend.sh` 已通过
- [ ] 已确认生产 `.env` 包含 `DATABASE_*`、`REDIS_*`、`JWT_SECRET`、`COS_SECRET_ID`、`COS_SECRET_KEY`
- [ ] 已轮换曾暴露的 COS 密钥
- [ ] 已确认 nginx upstream 指向宿主机 `3000` 端口

## 回滚信息

- 回滚目标版本: 上一版可用后端镜像或上一版 Git commit。
- 回滚步骤: 切回上一版代码后重新执行 `./deploy_backend.sh`，或手动运行上一版镜像。
- 回滚后验证: `curl http://127.0.0.1:3000/health` 返回成功，nginx 代理接口不再返回 502。
