# Product Taxonomy Structure Test Cases

## 测试范围

- Feature ID: SL-BE-PRODUCT-TAXONOMY-STRUCTURE
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 创建结构化商品 | 卖家账号具备 sales 能力 | POST `/products`，提交品种、牌号、钢厂、拆分规格和仓库信息 | 商品创建成功，状态为待审核，返回主数据 ID 和拆分规格 |
| TC-02 | 文本自动补主数据 | 请求不提交 `grade_id` / `steel_mill_id` | POST `/products`，只提交 `spec.material` 和 `steel_mill` | 后端自动创建或复用主数据，商品保存对应 ID |
| TC-03 | 重复捆包号 | 同企业或不同企业已有相同 `bundle_no` | 连续创建两个相同 `bundle_no` 商品 | 两个商品均可创建，不触发唯一约束 |
| TC-04 | 规格范围筛选 | 已有厚度和宽度结构化商品 | GET `/products?min_height_spec=3&max_height_spec=5` | 只返回命中规格范围的商品 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 绑定不存在的品种 ID | 请求提交无效 `category_id` | POST `/products` | 返回参数错误 |
| TC-EX-02 | 绑定不存在的牌号 ID | 请求提交无效 `grade_id` | POST `/products` | 返回参数错误 |
| TC-EX-03 | 绑定不存在的钢厂 ID | 请求提交无效 `steel_mill_id` | POST `/products` | 返回参数错误 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 未登录 | 创建商品 | 返回未登录 |
| TC-AUTH-02 | 无 sales 能力用户 | 创建商品 | 被 `RequireTradeCapability("sales")` 拦截 |
| TC-AUTH-03 | 公开用户 | 查询商品主数据 | 可读取启用的品种、牌号和钢厂 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 后端验证 | `go test ./...` | 单元测试通过 |
| TC-OPS-02 | 后端构建 | `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` | API 构建通过 |
| TC-OPS-03 | 发布门禁 | `./harness/release.sh docs/features/SL-BE-PRODUCT-TAXONOMY-STRUCTURE` | release gate 通过 |
