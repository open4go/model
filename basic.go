package model

import (
	"context"
	rtime "github.com/r2day/base/time"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// MetaModel 元模型
type MetaModel struct {
	// 上下文
	Context context.Context `json:"context" bson:"-"`
	// 数据库
	Handler *mongo.Database `json:"handler" bson:"-"`
	// 表名称
	Collection string `json:"collection" bson:"-"`
	// 商户号
	MerchantID string `json:"merchant_id" bson:"merchant_id"`
	// 创建者
	AccountID string `json:"account_id" bson:"account_id"`
	// 创建时间
	CreatedAt string `json:"created_at" bson:"created_at"`
	// 更新时间
	UpdatedAt string `json:"updated_at" bson:"updated_at"`
	// 状态
	Status bool `json:"status"`
	// 根据角色的最低级别写入
	AccessLevel uint `json:"access_level" bson:"access_level"`
}

// Model 模型
type Model struct {
	// 基本的数据库模型字段，一般情况所有model都应该包含如下字段
	// 创建时（用户上传的数据为空，所以默认可以不传该值)
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	// 基本的数据库模型字段，一般情况所有model都应该包含如下字段
	Meta MetaModel `json:"meta" bson:"meta"`
}

// Init 设置名称
func (m *Model) Init(ctx context.Context, handler *mongo.Database, name string) *Model {
	m.Meta.Context = ctx
	m.Meta.Handler = handler
	m.Meta.Collection = name
	return m
}

// Create 创建
func (m *Model) Create() (string, error) {
	// 保存时间设定
	m.Meta.CreatedAt = rtime.FomratTimeAsReader(time.Now().Unix())
	// 更新时间设定
	m.Meta.UpdatedAt = rtime.FomratTimeAsReader(time.Now().Unix())

	coll := m.Meta.Handler.Collection(m.Meta.Collection)
	// 插入记录
	result, err := coll.InsertOne(m.Meta.Context, m)
	if err != nil {
		log.WithField("m", m).Error(err)
		return "", err
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}
