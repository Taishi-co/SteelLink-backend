# SL-GOV-LARK-PR-DOC Test Cases

## TC1: 帮助信息

1. 执行 `./harness/lark-pr-doc.sh --help`
2. 确认输出包含 feature dir、PR URL、PR title 和 dry-run 参数说明

## TC2: Dry-run 不写入飞书

1. 执行 `./harness/lark-pr-doc.sh docs/features/SL-GOV-LARK-PR-DOC --pr-url https://github.com/example/pr/1 --pr-title "Example PR" --dry-run`
2. 确认输出包含 `create-doc` dry-run
3. 确认输出包含对共享 PR 记录 `Backend` 段落的 `update-doc` dry-run

## TC3: 缺少 feature 文档时失败

1. 指向一个不存在或缺少 `spec.md`、`test-cases.md`、`implementation-plan.md` 的 feature 目录
2. 确认脚本以非 0 状态退出，并输出缺失文件名
