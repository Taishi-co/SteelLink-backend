# Release Checklist

- [ ] `/cart/items` 返回 `image_url`。
- [ ] `/cart/items` 返回商品绑定仓库快照。
- [ ] 加入和更新购物车响应保持兼容。
- [ ] `CGO_ENABLED=0 go build -o /tmp/steellink-api ./cmd/api` 通过。
