# Implementation Plan

1. 检查订单详情和事件接口权限口径。
2. 在 `OrderService.GetByID` 的参与方判断中加入仓储企业匹配。
3. 保持 `ListEvents` 已有仓储方权限逻辑不变。
4. 运行 Go 测试、构建和 harness 验证。
