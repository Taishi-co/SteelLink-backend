# Test Cases

## 测试范围

- Feature ID: SL-BE-SERVICE-FULFILLMENT
- 测试负责人: Codex

## Happy Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-01 | 仓储资源 CRUD | 企业已开通 `warehouse` 服务 | POST 创建仓储资源，再 PUT 更新，再 DELETE 下线 | 返回资源详情，状态更新为已下线 |
| TC-02 | 物流资源 CRUD | 企业已开通 `logistics` 服务 | POST 创建物流资源，再 PUT 更新，再 DELETE 下线 | 返回资源详情，状态更新为已下线 |
| TC-03 | 商品绑定仓库 | 卖家拥有商品且仓储资源存在 | `PUT /products/:id/warehouse-binding` | 商品 `warehouse_info` 更新为平台登记仓库信息 |
| TC-04 | 查询仓库商品 | 仓库已有绑定商品 | `GET /warehouses/:id/products` | 返回该仓库绑定商品列表 |
| TC-05 | 创建出入库记录 | 仓储商已登录 | `POST /warehouse-movements` | 返回入库/出库记录，状态为待作业或请求状态 |
| TC-06 | 更新出入库状态 | 仓储商已登录且记录存在 | `PUT /warehouse-movements/:id/status` | 记录状态更新为作业中或已完成 |
| TC-07 | 物流价格估算 | 物流服务存在 | `POST /logistics/estimate` 提交重量、起终点 | 返回按吨价规则估算出的费用 |
| TC-08 | 买家发单物流商接单 | 买家已登录，物流服务存在 | POST 创建物流委托，再物流商 POST 接单 | 委托状态从待接单变为已接单 |
| TC-09 | 运输追踪 | 物流商已接单 | 更新物流单状态并追加轨迹节点 | 订单详情摘要可看到最新轨迹 |
| TC-10 | 履约保障 | 订单或物流单存在 | POST 创建保障记录并更新状态 | 返回保障类型、状态和处理说明 |

## Exception Path

| 编号 | 场景 | 前置条件 | 步骤 | 期望结果 |
| --- | --- | --- | --- | --- |
| TC-EX-01 | 非商品卖家绑定仓库 | 登录企业不是商品所属企业 | `PUT /products/:id/warehouse-binding` | 返回无权绑定商品仓库 |
| TC-EX-02 | 非物流服务商接单 | 企业未开通 `logistics` 服务 | `POST /logistics/orders/:id/accept` | 被 middleware 拦截 |
| TC-EX-03 | 不存在的物流服务估价 | 请求 `logistics_service_id` 不存在 | `POST /logistics/estimate` | 返回物流服务不存在 |

## 权限测试

| 编号 | 角色 | 场景 | 期望结果 |
| --- | --- | --- | --- |
| TC-AUTH-01 | 买家 | 创建物流委托、查看订单履约摘要 | 可访问采购侧物流履约接口 |
| TC-AUTH-02 | 卖家 | 绑定自己商品到仓库 | 可更新自己商品的仓库信息 |
| TC-AUTH-03 | 仓储商 | 管理仓库资源和出入库记录 | 可访问仓储作业接口 |
| TC-AUTH-04 | 物流商 | 接单、更新运输、追加轨迹 | 可访问物流履约接口 |

## 配置与部署验证

| 编号 | 场景 | 验证命令/步骤 | 期望结果 |
| --- | --- | --- | --- |
| TC-OPS-01 | 配置文件可加载 | `./harness/verify.sh` | Go 构建通过，配置解析无阻断 |
