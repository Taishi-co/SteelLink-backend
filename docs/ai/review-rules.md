# Review Rules

Review 不是只看代码风格，而是确认“实现是否符合设计、验证是否足够、上线风险是否可控”。

## Review 输入

发起 review 前必须提供：

- feature 文档目录
- `spec.md`
- `test-cases.md`
- `implementation-plan.md`
- `./harness/verify.sh` 的结果

## 阻断项

- 缺少 feature 文档
- `go test ./...` 未通过
- `CGO_ENABLED=0 go build -o <tmp-output> ./cmd/api` 未通过
- 需求实现与 `spec.md` 不一致
- 测试用例未覆盖主要 happy path、异常 path、权限 path
- 涉及权限、审核或统计口径改动但未核对角色矩阵
- 涉及发布或环境变量改动但未更新 `release-checklist.md`

## Review 关注点

- 是否改动了错误的 Gin / handler / middleware / pkg 分层边界
- 是否遗漏权限控制、幂等、日志留痕或审计点
- 是否引入管理后台不可观测的风险
