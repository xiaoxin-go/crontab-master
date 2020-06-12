package controllers

import (
	"context"
	"crontab/common"
	"crontab/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogController struct{
	utils.HttpController
}

func (this *LogController) GetList(){
	var (
		page int64
		pageSize int64
		err error
	)
	name := this.GetString(":name")
	if page, err = this.GetInt64("page"); err != nil{
		page = 1
	}
	if pageSize, err = this.GetInt64("page_size"); err != nil{
		pageSize = 20
	}

	filter := &common.JobLogFilter{JobName: name}
	logSort := &common.SortLogByStartTime{SortOrder: -1}

	findOptions := options.Find()
	findOptions.SetSort(logSort)
	findOptions.SetSkip((page-1)*pageSize)
	findOptions.SetLimit(pageSize)
	collection := utils.G_mongoClient.GetCollection("cron", "log")
	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	this.HttpServerError(err, "查询日志异常")

	result := make([]*common.JobLog, 0)
	err = utils.G_mongoClient.ReadAll(cursor, &result)
	this.HttpServerError(err, "查询日志读取异常")
	this.HttpSuccess(result, "ok")
}
