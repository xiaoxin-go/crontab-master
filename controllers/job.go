package controllers

import (
	"crontab/common"
	"crontab/utils"
	"encoding/json"
)

type JobController struct{
	utils.HttpController
}

// 获取任务列表
func (this *JobController) GetList(){
	// 获取任务名，拼接
	name := this.GetString("name")
	name = common.JOB_SAVE_DIR + name

	// 从etcd中获取任务列表
	result, err := utils.G_EtcdClient.GetList(name)
	this.HttpServerError(err, "获取数据异常")
	// 数据获取成功，将数据转换成job格式返回
	jobList := make([]*common.Job, 0)
	for _, jobBytes := range result{
		job := common.Job{}
		err = json.Unmarshal(jobBytes, &job)
		this.HttpServerError(err, "数据转换异常")
		jobList = append(jobList, &job)
	}
	this.HttpSuccess(jobList, "ok")
}

// 新建或修改任务
func (this *JobController) Save(){
	// 获取参数
	job := common.Job{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &job)
	this.HttpParamsError(err, "参数异常")

	// 组合key
	key := common.JOB_SAVE_DIR + job.Name

	// 写入数据，返回
	_, err = utils.G_EtcdClient.Put(key, job)
	this.HttpServerError(err, "保存数据异常")
	this.HttpSuccess(nil, "ok")
}

// 删除任务
func (this *JobController) Del(){
	name := this.GetString(":name")
	name = common.JOB_SAVE_DIR + name
	_, err := utils.G_EtcdClient.Delete(name)
	this.HttpServerError(err, "删除ETCD数据异常")
	this.HttpSuccess(nil, "ok")
}

// 杀死任务
func (this *JobController) Kill(){
	// 获取任务名，往Etcd写入数据，创建一个租约，让其自动过期
	name := this.GetString(":name")
	name = common.JOB_KILLER_DIR + name

	_, err := utils.G_EtcdClient.PutLease(name, "", 10)
	this.HttpServerError(err, "写入ETCD异常")
	this.HttpSuccess(nil, "操作成功")
}
