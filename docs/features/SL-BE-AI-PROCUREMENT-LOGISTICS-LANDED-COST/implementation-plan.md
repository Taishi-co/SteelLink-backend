# SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST Implementation Plan

## Feature 信息

- Feature ID: SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST
- 开发负责人: Codex

## 变更范围

- 代码文件:
  - `internal/modules/aiprocurement/domain/entity/procurement.go`
  - `internal/modules/aiprocurement/application/service/service.go`
  - `internal/modules/aiprocurement/ui/dto/dto.go`
  - `internal/modules/aiprocurement/ui/handler/handler.go`
  - `internal/modules/aiprocurement/infrastructure/router/router.go`
  - `internal/modules/aiprocurement/application/service/model_parser.go`
  - `internal/modules/aiprocurement/application/service/model_parser_test.go`
- 文档文件:
  - `docs/features/SL-BE-AI-PROCUREMENT-LOGISTICS-LANDED-COST/*`
- 配置文件: 无
- 部署/脚本文件: 无

## 实施步骤

1. 扩展 AI采购解析参数和匹配商品 JSON 字段，保持旧数据兼容。
2. 在商品匹配和订单生成中间新增 `logistics_warehouse_matching` 工作流阶段与 `match-logistics-warehouse` 接口。
3. 商品匹配阶段只返回商品候选；物流仓储阶段加载已审核物流商，用商品仓库地址、收货地址、线路、服务范围、载重能力筛选多个候选物流方案。
4. 无适合物流方案时返回 `no_match`，有多个方案时默认选最优方案，同时保留 `logistics_options` 供用户改选。
5. 在生成订单前按最终拼单数量和用户选择的物流方案重新估算运费，协商物流状态不被重算覆盖。
6. 生成真实订单时按卖家和发货仓库分组，确保同一真实订单只包含同一发货仓商品。
7. 更新模型解析提示词，使大模型能抽取收货地址。

## 验证计划

- `gofmt`
- `go test ./...`
- `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api`
- `./harness/verify.sh`
- 需要的人工验证: 使用买家账号在 `/ai-procurement` 输入收货地址，确认商品匹配后新增物流仓储匹配步骤并展示到手价。

## 风险与回滚关注点

- 运费估算当前基于物流商价格规则中的首个数字作为元/吨费率，属于平台内部估算，不替代履约阶段正式报价。
- 回滚时删除本 feature 分支改动即可，JSON 新字段为空不会影响旧流程。
