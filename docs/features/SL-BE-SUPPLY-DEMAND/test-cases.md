# Test Cases: 供求信息后端

## 测试范围

- Feature ID: SL-BE-SUPPLY-DEMAND
- 测试负责人: AI agent

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 公开列表 | 存在已发布供求信息 | `GET /api/v1/supply-demands` | 返回分页列表，仅包含已发布信息 |
| TC-02 | 公开详情 | 存在已发布供求信息 | `GET /api/v1/supply-demands/:id` | 返回供求信息详情 |
| TC-03 | 发布信息 | 登录用户 | `POST /api/v1/supply-demands` 提交 `supply` 或 `demand` | 创建成功并返回详情 |
| TC-04 | 我的发布 | 登录用户已发布信息 | `GET /api/v1/supply-demands/mine` | 返回当前用户发布的信息 |
| TC-05 | 状态更新 | 登录用户为发布人 | `PUT /api/v1/supply-demands/:id/status` | 状态更新成功 |
| TC-06 | 统计信息 | 任意用户 | `GET /api/v1/supply-demands/stats` | 返回已发布信息数、撮合成功数、撮合金额数、今日新增数 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 未登录发布 | 游客 | `POST /api/v1/supply-demands` | 返回未登录错误 |
| TC-EX-02 | 类型非法 | 登录用户 | `POST /api/v1/supply-demands` 提交非法 `type` | 返回参数错误 |
| TC-EX-03 | 非发布人更新状态 | 登录用户不是发布人 | `PUT /api/v1/supply-demands/:id/status` | 返回无权限错误 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 游客 | 查询列表/详情/统计 | 可访问 |
| TC-AUTH-02 | 登录用户 | 发布/我的发布/状态更新 | 可访问自己的数据 |
| TC-AUTH-03 | 登录用户 | 更新他人发布状态 | 被拒绝 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | Go 测试 | `go test ./...` | 通过 |
| TC-OPS-02 | Go 构建 | `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` | 通过 |
| TC-OPS-03 | Harness 验证 | `./harness/review.sh docs/features/SL-BE-SUPPLY-DEMAND` | 通过 |
