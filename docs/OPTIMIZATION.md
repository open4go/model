# model 优化方案

> 依据：`basic.go`、`util.go`、`demo.go`、CI 配置、`govulncheck` 结果（2026-10-22）。工作量：S ≤ 0.5 天，M 1–2 天，L ≥ 3 天。

| # | 优先级 | 问题 | 位置 | 建议 | 工作量 |
|---|---|---|---|---|---|
| 1 | 高 | 非法 id 被静默忽略：`ObjectIDFromHex` 的错误被丢弃，非法 id 变成 `NilObjectID`；`Delete` 匹配不到也返回 nil，调用方以为删除成功 | `basic.go:168,187,212,319,387`（Delete/GetOne/Update/SoftDelete/Transfer） | 统一返回 `ErrInvalidID`；`Delete` 在 `DeletedCount==0` 时返回 `mongo.ErrNoDocuments` | S |
| 2 | 高 | `Update`、`SoftDelete` 在 `MatchedCount<1` 时 `return err`，而此时 err 恒为 nil，“更新不存在的记录”被当成成功 | `basic.go` Update / SoftDelete | 返回明确错误（与 `UpdateV2` 保持一致） | S |
| 3 | 高 | `UpdateMany` 把自定义字段写到 `$set.data` 下，而不是合并到顶层，调用方以为更新了字段，实际只多出一个 `data` 字段 | `basic.go:368` | 改成 `mergeBsonM(meta, updateData)`；如果确实需要写到 `data`，在命名和注释里说清楚 | S |
| 4 | 中 | `Create` 对 `InsertedID` 做强制类型断言，自定义 `_id`（字符串等）时会 panic | `basic.go:157` | 用 `id, ok := ...(primitive.ObjectID)`，不 ok 时用 `fmt.Sprint` | S |
| 5 | 中 | `GetList` 的计数用 `context.TODO()`，没用请求上下文，超时和取消都不生效；`CountDocuments` 的其他错误被忽略 | `basic.go:266` | 用 `m.Context.Context`，非 nil 错误直接返回 | S |
| 6 | 中 | context key 用裸字符串（`"MERCHANT_KEY"` 等），容易和其他包冲突（staticcheck SA1029） | `basic.go` 常量、`util.go` | 定义私有 `type ctxKey string`，提供 `WithMerchant(ctx, id)` 等 setter；过渡期同时读新旧两种 key（参考 `log/context.go` 的做法） | M |
| 7 | 中 | `Transfer` 修改的是 `m.Meta`，但写回的是 `d`，只有 `d` 正好嵌入了同一个 `*Model` 才会生效，很脆弱 | `basic.go` Transfer | 改为显式 `$set: {"meta.merchant_id": merchantId, ...}` | S |
| 8 | 中 | 嵌入方式不统一：`auth` 用 `bson:"_"`，`log/model/operation` 不加 tag（驱动按小写字段名 `model` 存）；`rest` 的租户过滤又写死 `_.meta.merchant_id`。各集合的元数据路径不同，`rest` 的通用过滤对部分集合不起作用 | 跨仓库 | 在 README 里定下统一约定（推荐 `bson:",inline"` 或统一 `_`），并提供 `MetaPath()` 常量给 `rest` 使用 | M |
| 9 | 中 | 没有任何单元测试 | 全仓库 | 用 `mtest` 或 testcontainers 覆盖 CRUD、Meta 自动填充、软删除 | M |
| 10 | 低 | `demo.go` 放在库包里，引入了 `r2day/db` 依赖，还会 panic | `demo.go` | 改成 `example_test.go`（`ExampleModel_Create`），去掉 `r2day/db` 依赖 | S |
| 11 | 低 | `SoftDelete` 用 `time.Now().Unix()`，其他方法用 `r3time.CurrentTimestamp()`，不一致 | `basic.go` SoftDelete | 统一用 `r3time` | S |
| 12 | 中 | 依赖漏洞：`golang.org/x/net v0.29.0`（GO-2026-6617/6612/6611/6603）；通过旧版 `log v0.0.16` 间接引入 `github.com/docker/docker`（GO-2026-4887/4883） | `go.mod` | 升级 `log` 到 v0.0.23+（新版已去掉 docker 依赖），`go get golang.org/x/net@latest` | S |
| 13 | 低 | 与 `log` 循环依赖（model→log v0.0.16，log→model v0.0.4） | `go.mod` | 把 `log/model/*` 移到独立模块或业务仓库，让 `log` 不再依赖 `model` | M |
| 14 | 低 | CI：govulncheck 失败导致整个流水线红；golangci-lint 固定在 v1.55.2 | `.github/workflows/go.yml` | 先修依赖；govulncheck 单独成 job；升级 lint 到 v2 | S |
