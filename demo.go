package model

import (
	"context"
	"fmt"
	"github.com/r2day/db"
)

type DemoModel struct {
	Model
	// 名称
	Name string `json:"name" bson:"name"`
	// 描述
	Desc string `json:"desc" bson:"desc"`
	// 引用次数
	Reference uint `json:"reference" bson:"reference"`
}

func Demo() {
	d := &DemoModel{}
	m := d.Init(context.TODO(), db.MDB, "demo")
	s, err := m.Create()
	if err != nil {
		panic(err)
	}
	fmt.Println(s)
}
