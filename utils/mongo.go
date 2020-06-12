package utils

import (
	"context"
	"crontab/conf"
	"encoding/json"
	"github.com/astaxie/beego"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


// 连接mongodb数据库
type MongoClient struct{
	client *mongo.Client
	collection *mongo.Collection
}

var (
	G_mongoClient *MongoClient
)

// 初始化连接mongo
func init(){
	clientOptions := options.Client().ApplyURI(conf.MongodbUri)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil{
		panic(err)
	}
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil{
		beego.Info("mongo数据库连接异常")
		panic(err)
	}
	G_mongoClient = &MongoClient{client: client}
	return
}

// 设置mongo db table
func (this *MongoClient)GetCollection(db, name string)*mongo.Collection{
	return this.client.Database(db).Collection(name)
}

// 读取查询过的数据
func (this *MongoClient)ReadAll(cursor *mongo.Cursor, result interface{})(err error){
	itemList := make([]interface{}, 0)
	for cursor.Next(context.TODO()){
		item := map[string]interface{}{}
		err = cursor.Decode(&item)
		if err != nil{
			return
		}
		itemList = append(itemList, item)
	}
	bytes, err := json.Marshal(itemList)
	if err != nil{
		return
	}
	err = json.Unmarshal(bytes, result)
	return
}



