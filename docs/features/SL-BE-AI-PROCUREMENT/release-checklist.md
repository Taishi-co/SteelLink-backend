# SL-BE-AI-PROCUREMENT Release Checklist

- [ ] `gofmt` 已执行。
- [ ] `go test ./...` 通过。
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` 通过。
- [ ] `./harness/review.sh docs/features/SL-BE-AI-PROCUREMENT` 通过。
- [ ] `./harness/release.sh docs/features/SL-BE-AI-PROCUREMENT` 通过。
- [ ] 前端联调确认五阶段数据结构可渲染。
