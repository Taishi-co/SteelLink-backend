# Test Cases

| ID | 场景 | 前置条件 | 操作 | 期望结果 |
|----|------|----------|------|----------|
| TC-01 | 物流确认报价创建合同 | 订单 status=16，竞价 open | 物流方 POST /logistics-confirm | Order → 27，创建 LogisticsContract(pending) |
| TC-02 | 买方签署物流合同 | Order status=27，合同 pending | 买方 POST /logistics-contracts/buyer-sign | 合同 → buyer_signed，Order 仍为 27 |
| TC-03 | 物流方签署物流合同（双方签完） | 合同 buyer_signed | 物流方 POST /logistics-contracts/logistics-sign | 合同 → completed，Order → 18，创建 ShipmentTask |
| TC-04 | 签署顺序反转 | 合同 pending | 物流先签 → 买方再签 | 同样 → completed → Order 18 |
| TC-05 | 非融资订单确认收货创建结算单 | Order status=7，无融资 | 买方 POST /receipts/confirm | Order → 19，创建 LogisticsSettlement |
| TC-06 | 结算单金额计算 | 运费 6000，抽佣率 3% | 查看结算单 | 抽佣=180，实付=5820 |
| TC-07 | 支付物流费 | 结算 pending | POST /logistics-settlements/pay | 结算 → paid |
| TC-08 | 支付平台抽佣 | 结算 paid | POST /logistics-settlements/commission | 结算 → committed |
| TC-09 | 上传履约单据 | 物流委托单存在 | POST /logistics/orders/:id/documents | 创建 FulfillmentDocument |
| TC-10 | 铅封确认 | 物流委托单存在 | POST /logistics/orders/:id/seal-confirm | 创建 seal_photo 文档 + tracking event |
| TC-11 | 验货结果 | 物流委托单存在 | POST /logistics/orders/:id/inspection | 创建 inspection_result + tracking event |
| TC-12 | 融资订单 is_financing_order 标记 | 订单 finance_id > 0 | 物流确认报价 | LogisticsContract.IsFinancingOrder = true |
| TC-13 | 融资路径资方确认收货 | 融资订单 status=7 | 资方 POST /receipts/confirm | Order → 19（资方有权） |
| TC-14 | 目的地仓确认权限 | dest_warehouse_enterprise_id=5 | 企业5 POST /warehouse-confirm | 确认成功 |
| TC-15 | 提货仓无权确认 | warehouse_enterprise_id=4 | 企业5 POST /warehouse-confirm | 403 无权 |
| TC-16 | fee_voucher 上传 | 上传手续费凭证 | POST /upload purpose=fee_voucher | 上传成功 |
| TC-17 | 融资 timeline 展开 | 订单已放款 | GET /orders/:id | timeline 包含风控→合同→付费→放款→质押步骤 |
