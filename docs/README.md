# model — MongoDB 通用数据模型基类

> 仓库：https://github.com/open4go/model ｜ 模块路径：`github.com/open4go/model` ｜ Go 1.24.0 ｜ 最近提交：2026-10-09（最新 tag v0.0.25）

## 项目简介
`model` 给业务服务提供一个可嵌入的 `Model` 基类，封装 MongoDB 单集合 CRUD（创建、查询、更新、软删除、批量更新、数据转移），写入时自动从 `context` 里取命名空间、商户、账号、操作人等元数据，填到 `MetaModel`。r2day/* 业务服务、`auth`、`log` 等仓库的数据模型都嵌入它。

## 目录结构
```
model/
├── basic.go   # MetaModel / MetaContext / Model 定义与全部 CRUD 方法
├── util.go    # GetValueFromCtx：从 context 读取字符串值
├── tool.go    # mergeBsonM：合并两个 bson.M
├── demo.go    # DemoModel 示例（依赖 github.com/r2day/db 全局 MDB）
└── .github/workflows/go.yml  # govulncheck + golangci-lint + build + test
```

## 核心类型与对外 API
| 名称 | 说明 |
|---|---|
| 常量 `AccountKey` / `MerchantKey` / `NamespaceKey` / `OperatorKey` / `OldMerchantKey` | context key（字符串类型），由 `middle` 写入，`model` 读取 |
| `MetaModel` | 元数据：namespace、merchant_id、old_merchant_id、founder、updater、account_id、created_at/updated_at（字符串）、created_time/updated_time（时间戳）、status、deleted、access_level |
| `MetaContext` | 运行时上下文：`Context`、`Handler *mongo.Database`、`Collection`（均 `bson:"-"`） |
| `Model` | `Meta MetaModel` + `Context MetaContext` |
| `(*Model).Init(ctx, db, name) *Model` / `NewModel(...)` | 绑定上下文、库和集合名 |
| `Create(d) (string, error)` | 生成 Meta 后 `InsertOne`，返回 ObjectID 十六进制 |
| `GetOne(d, id)` / `GetBy(d, filter)` | 按 `_id` 或自定义 filter 查询单条 |
| `GetList(filter, d)` / `GetListWithOpt(filter, d, opt)` | `CountDocuments` + `Find`，返回总数 |
| `Update(d, id)` / `UpdateV2(bson.M, id)` | `$set` 更新；V2 自动合并 `meta.updater/updated_at/updated_time` |
| `UpdateMany(filter, updateData)` | 批量更新（见优化方案：自定义字段被写到 `data` 字段下） |
| `Delete(id)` / `SoftDelete(id)` | 物理删除 / 标记 `Meta.Deleted=true` |
| `Transfer(d, id, merchantId)` | 把数据转给另一个商户 |
| `GetValueFromCtx(ctx, key) string` | 取不到或不是字符串时返回 "" |

## 使用示例（取自仓库中真实代码）
```go
// demo.go
type DemoModel struct {
    Model
    Name string `json:"name" bson:"name"`
}
d := &DemoModel{Name: "i am demo"}
m := d.Init(context.TODO(), db.MDB, "demo")
id, err := m.Create(d)
```
`auth/app.go` 中的列表查询：
```go
m := &app.Model{}
apps := make([]app.Model, 0)
handler := m.Init(context.TODO(), GetDBHandler(a.Ctx), m.CollectionName())
_, err := handler.GetList(bson.D{}, &apps)
```

## 依赖
- `go.mongodb.org/mongo-driver v1.17.3`
- `github.com/open4go/log v0.0.16`、`github.com/open4go/r3time v0.0.6`
- `github.com/r2day/db v0.3.5`（只有 `demo.go` 用到）

## 构建与使用
```bash
go get github.com/open4go/model@latest
go build ./... && go vet ./...   # 本次检查：均通过
```
仓库没有测试文件。CI 最近几次都失败在 `Run govulncheck` 这一步。

## 与其他仓库的关系
- 被依赖：`log`（`log/model/*` 嵌入 `model.Model`）、`auth`（account/app/role/tenant 模型）、`middle`（读写 context key）、`rest`（间接）、r2day/* 业务服务。
- 依赖：`log`、`r3time`。**`log` 和 `model` 互相依赖**（版本不同），形成模块级循环。
- 约定：`middle.WriteIntoHeader` 写入 `MerchantKey/AccountKey/OperatorKey/NamespaceKey`，`Model.GetMeta()` 读取这些 key。
