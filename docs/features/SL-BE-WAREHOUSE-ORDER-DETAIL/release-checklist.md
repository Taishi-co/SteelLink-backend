# Release Checklist

- [x] 已从最新 `main` 创建修复分支。
- [x] `go test ./...` 通过。
- [x] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` 通过。
- [x] `./harness/verify.sh` 通过。
- [ ] 回滚方式：回退 `internal/modules/order/application/service/order_service.go` 的详情权限条件。
