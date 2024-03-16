package model

import (
	"context"
	"errors"
	"github.com/open4go/r3time"
	rtime "github.com/r2day/base/time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	AccountKey   = "ACCOUNT_KEY"
	MerchantKey  = "MERCHANT_KEY"
	NamespaceKey = "NAME_SPACE_KEY"
	OperatorKey  = "OPERATOR_KEY"
)

// MetaModel 元模型
type MetaModel struct {
	// 命名空间 (例如：集团名）
	Namespace string `json:"namespace" bson:"namespace"`
	// 商户号 （例如：组织，分公司id）
	MerchantID string `json:"merchant_id" bson:"merchant_id"`
	// 创建者 （具体的数据创建人） updater
	Founder string `json:"founder" bson:"founder"`
	// 更新人 （具体的数据创建人） updater
	Updater string `json:"updater" bson:"updater"`
	// 数据所属人
	AccountID string `json:"account_id" bson:"account_id"`
	// 创建时间
	CreatedAt string `json:"created_at" bson:"created_at"`
	// 更新时间
	UpdatedAt string `json:"updated_at" bson:"updated_at"`
	// 创建时间 （时间戳)
	CreatedTime int64 `json:"created_time" bson:"created_time"`
	// 更新时间 （时间戳)
	UpdatedTime int64 `json:"updated_time" bson:"updated_time"`
	// 状态
	Status bool `json:"status"`
	// 如果数据被删除则直接标识为删除 （不进行物理删除）
	Deleted bool `json:"deleted"`
	// 根据角色的最低级别写入
	AccessLevel uint `json:"access_level" bson:"access_level"`
}

type MetaContext struct {
	// 上下文
	Context context.Context `json:"-" bson:"-"`
	// 数据库
	Handler *mongo.Database `json:"-" bson:"-"`
	// 表名称
	Collection string `json:"-" bson:"-"`
}

// CopyMeta copies the meta data from src to dest
func CopyMeta(src, dest *Model) {
	dest.Meta = src.Meta
}

// Model 模型
type Model struct {
	// 基本的数据库模型字段，一般情况所有model都应该包含如下字段
	Meta MetaModel `json:"meta" bson:"meta"`
	// 基本的数据库模型字段，一般情况所有model都应该包含如下字段
	Context MetaContext `json:"context" bson:"context"`
}

// Init 设置名称
// 当需要执行父类方法时可以直接使用返回的handler完成调用
// 例如:
// handler := m.Init(c.Request.Context(), store.MongoDatabase, m.CollectionName())
// handler.Create(m)
// 如果用户希望执行自己定义的特殊的method则需要进行handler context 复制以便进行子类方法的运行
//
//	handler := m.Init(c.Request.Context(), store.MongoDatabase, m.CollectionName())
//	m.Meta = handler.Meta
//	_, err := m.GetListDesc(bson.D{
//		{Key: "basic.meta.account_id", Value: accountID},
//		{"order.status", bson.D{{"$lte", 4}}},
//	}, &menuList)
func (m *Model) Init(ctx context.Context, handler *mongo.Database, name string) *Model {
	m.Context.Context = ctx
	m.Context.Handler = handler
	m.Context.Collection = name
	return m
}

// NewModel 创建新模型
func NewModel(ctx context.Context, handler *mongo.Database, name string) *Model {
	m := &Model{}
	m.Context.Context = ctx
	m.Context.Handler = handler
	m.Context.Collection = name
	return m
}

// Create 创建
func (m *Model) Create(d interface{}) (string, error) {
	// 保存时间
	m.Meta.CreatedTime = time.Now().Unix()
	// 更新时间
	m.Meta.UpdatedTime = time.Now().Unix()
	// 创建时间
	m.Meta.CreatedAt = r3time.CurrentTime()
	// 更新时间
	m.Meta.UpdatedAt = r3time.CurrentTime()
	// 命名空间
	m.Meta.Namespace = GetValueFromCtx(m.Context.Context, NamespaceKey)
	// 商户
	m.Meta.MerchantID = GetValueFromCtx(m.Context.Context, MerchantKey)
	// 数据操作所属人
	m.Meta.AccountID = GetValueFromCtx(m.Context.Context, AccountKey)
	// 创建人
	m.Meta.Founder = GetValueFromCtx(m.Context.Context, OperatorKey)
	// 更新人
	m.Meta.Updater = GetValueFromCtx(m.Context.Context, OperatorKey)

	// Copy meta data from context
	CopyMeta(m, d.(*Model))

	coll := m.Context.Handler.Collection(m.Context.Collection)
	// 插入记录
	result, err := coll.InsertOne(m.Context.Context, d)
	if err != nil {
		return "", err
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}

// Delete 删除
// delete	DELETE http://my.api.url/posts/123
func (m *Model) Delete(id string) error {
	// 更新时间设定
	m.Meta.UpdatedAt = rtime.FomratTimeAsReader(time.Now().Unix())

	coll := m.Context.Handler.Collection(m.Context.Collection)
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	// 执行删除
	result, err := coll.DeleteOne(m.Context.Context, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount < 1 {
		return nil
	}
	return nil
}

// GetOne 详情
// getOne	GET http://my.api.url/posts/123
func (m *Model) GetOne(d interface{}, id string) error {
	coll := m.Context.Handler.Collection(m.Context.Collection)
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	err := coll.FindOne(m.Context.Context, filter).Decode(d)
	if err != nil {
		return err
	}
	return nil
}

// GetBy 通过自定义查询字段
// getOne	GET http://my.api.url/posts/123
func (m *Model) GetBy(d interface{}, filter interface{}) error {
	coll := m.Context.Handler.Collection(m.Context.Collection)
	err := coll.FindOne(m.Context.Context, filter).Decode(d)
	if err != nil {
		return err
	}
	return nil
}

// Update 更新
// update PUT http://my.api.url/posts/123
// 会自动更新操作人与操作时间
func (m *Model) Update(d interface{}, id string) error {
	coll := m.Context.Handler.Collection(m.Context.Collection)
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	// 更新人
	m.Meta.Updater = GetValueFromCtx(m.Context.Context, OperatorKey)
	// 更新时间
	m.Meta.UpdatedTime = time.Now().Unix()
	// 更新时间
	m.Meta.UpdatedAt = r3time.CurrentTime()

	result, err := coll.UpdateOne(m.Context.Context, filter, bson.D{{Key: "$set", Value: d}})
	if err != nil {
		return err
	}

	if result.MatchedCount < 1 {
		return err
	}
	return nil
}

// GetList 获取列表
// getList	GET http://my.api.url/posts?sort=["title","ASC"]&range=[0, 24]&filter={"title":"bar"}
func (m *Model) GetList(filter interface{}, d interface{}) (int64, error) {
	coll := m.Context.Handler.Collection(m.Context.Collection)
	// 声明需要返回的列表
	//results := make([]*Model, 0)
	// 获取总数（含过滤规则）
	totalCounter, err := coll.CountDocuments(context.TODO(), filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, err
	}
	// 获取数据列表
	cursor, err := coll.Find(m.Context.Context, filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return totalCounter, err
	}

	if err != nil {
		return totalCounter, err
	}

	if err = cursor.All(m.Context.Context, d); err != nil {
		return totalCounter, err
	}
	return totalCounter, nil
}

// GetListWithOpt 获取列表
// GetListWithOpt	GET http://my.api.url/posts?sort=["title","ASC"]&range=[0, 24]&filter={"title":"bar"}
func (m *Model) GetListWithOpt(filter interface{}, d interface{}, opt *options.FindOptions) (int64, error) {
	coll := m.Context.Handler.Collection(m.Context.Collection)
	// 声明需要返回的列表
	//results := make([]*Model, 0)
	// 获取总数（含过滤规则）
	totalCounter, err := coll.CountDocuments(context.TODO(), filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, err
	}
	// 获取数据列表
	cursor, err := coll.Find(m.Context.Context, filter, opt)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return totalCounter, err
	}

	if err != nil {
		return totalCounter, err
	}

	if err = cursor.All(m.Context.Context, d); err != nil {
		return totalCounter, err
	}
	return totalCounter, nil
}

// SoftDelete 软删除
// update PUT http://my.api.url/posts/123
// 会自动更新操作人与操作时间
func (m *Model) SoftDelete(id string) error {

	coll := m.Context.Handler.Collection(m.Context.Collection)
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	// 查找当前数据库中的真实值
	err := coll.FindOne(m.Context.Context, filter).Decode(m)
	if err != nil {
		return err
	}

	// 更新人
	m.Meta.Updater = GetValueFromCtx(m.Context.Context, OperatorKey)
	// 更新时间
	m.Meta.UpdatedTime = time.Now().Unix()
	// 更新时间
	m.Meta.UpdatedAt = r3time.CurrentTime()
	// 标识为删除
	m.Meta.Deleted = true

	// 重新更新
	result, err := coll.UpdateOne(m.Context.Context, filter, bson.D{{Key: "$set", Value: m}})
	if err != nil {
		return err
	}

	if result.MatchedCount < 1 {
		return err
	}
	return nil
}
