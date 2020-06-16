package worker

import (
	"context"
	"crontab/worker/common"
	"crontab/worker/utils"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"strings"
)

type JobMgr struct{

}

var (
	// 单例
	G_jobMgr *JobMgr
)

// 监听任务变化
func (jobMgr *JobMgr) watchJobs()(err error){
	var jobEvent *common.JobEvent

	// 1. 获取/cron/jobs/目录下所有任务，并且获取当前集群的revision
	getResp, err := utils.G_EtcdClient.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
	if err != nil{
		return
	}
	for _, kvPair := range getResp.Kvs{
		// 反序列化json得到job
		job := common.Job{}
		err := json.Unmarshal(kvPair.Value, &job)
		if err != nil{
			continue
		}
		jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, &job)

		// TODO： 把jobEvent同步给scheduler调度协程
	}

	// 2. 从该revision向后监听变化事件
	go func(){
		// 从get时刻的后续版本开始监听变化
		watchStartRevision := getResp.Header.Revision + 1

		// 启动监听/cron/jobs/目录的后续变化
		watcher := utils.G_EtcdClient.Watcher()
		watchChan := watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		// 处理监听事件
		for watchResp := range watchChan{
			for _, watchEvent := range watchResp.Events{
				switch watchEvent.Type{
				case mvccpb.PUT:	// 任务保存事件
					// TODO: 反序列化Job, 推一个更新事件给Scheduler
					job := common.Job{}
					err := json.Unmarshal(watchEvent.Kv.Value, &job)
					if err != nil{
						continue
					}
					// 构建一个Event事件
					jobEvent = &common.JobEvent{
						EventType: common.JOB_EVENT_SAVE,
						Job: &job,
					}
				case mvccpb.DELETE: // 任务删除事件
					// 任务删除事件
					// TODO: 推一个删除事件给scheduler
					// 提取任务名
					jobName := strings.TrimPrefix(string(watchEvent.Kv.Key), common.JOB_SAVE_DIR)
					job := &common.Job{
						Name: jobName,
					}

					// 构建一个删除Event
					jobEvent = &common.JobEvent{
						EventType: common.JOB_EVENT_DELETE,
						Job: job,
					}
				}
				// TODO: 调度协程，推给Scheduler
				// G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()

	return
}

// 监听强杀任务通知
func (jobMgr *JobMgr) watchKiller(){
	// 监听/cron/killer目录
	go func(){
		watch := utils.G_EtcdClient.Watcher()
		watchChan := watch.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		for watchResp := range watchChan{
			for _, watchEvent := range watchResp.Events{
				switch watchEvent.Type{
				case mvccpb.PUT:
					jobName := strings.TrimPrefix(string(watchEvent.Kv.Key), common.JOB_KILLER_DIR)
					job := &common.Job{Name: jobName}
					jobEvent := &common.JobEvent{EventType: common.JOB_EVENT_KILL, Job: job}

					// 把事件推给scheuler
					// G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
				}
			}
		}
	}()
}

// 初始化管理器
func InitJobMgr()(err error){
	G_jobMgr = &JobMgr{}
	// 启动任务监听
	err = G_jobMgr.watchJobs()
	if err != nil{
		return
	}

	// 启动监听killer
	G_jobMgr.watchKiller()
	return
}

// 创建任务执行锁
func (job *JobMgr)CreateJobLock(jobName string)(jobLock *JobLock){
	jobLock = InitJobLock(jobName)
	return
}