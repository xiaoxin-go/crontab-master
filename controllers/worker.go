package controllers

import (
	"crontab/common"
	"crontab/utils"
)

type WorkController struct{
	utils.HttpController
}

func (this *WorkController) GetList(){
	name := this.GetString("name")
	name = common.JOB_WORKER_DIR + name

	// 从etcd中获取任务列表
	result, err := utils.G_EtcdClient.GetList(name)
	this.HttpServerError(err, "获取数据异常")
	// 数据获取成功，将数据转换成job格式返回
	workerList := make([]string, 0)
	for key, _ := range result{
		workerList = append(workerList, key)
	}
	this.HttpSuccess(workerList, "ok")
}