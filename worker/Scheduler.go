package worker

import (
	"crontab/worker/common"
	"fmt"
	"time"
)

// 任务调度
type Scheduler struct{
	jobEventChan chan *common.JobEvent
	jobPlanTable map[string]*common.JobSchedulerPlan
	jobExecutingTable map[string]*common.JobExecuteInfo
	jobResultChan chan *common.JobExecuteResult
}

var (
	G_scheduler *Scheduler
)

// 处理任务事件
func (scheduler *Scheduler)handleJobEvent(jobEvent *common.JobEvent){
	var (
		jobSchedulerPlan *common.JobSchedulerPlan
		jobExisted bool
		jobExecuting bool
		jobExecuteInfo *common.JobExecuteInfo
		err error
	)
	switch jobEvent.EventType{
	case common.JOB_EVENT_SAVE:	// 保存任务，将任务添加到执行计划表中
		if jobSchedulerPlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil{
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulerPlan
	case common.JOB_EVENT_DELETE:		// 删除任务，将任务从执行计划表中删除
		if jobSchedulerPlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted{
			delete(scheduler.jobPlanTable, jobSchedulerPlan.Job.Name)
		}
	case common.JOB_EVENT_KILL:			// 杀死任务，执行取消函数，杀死任务
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting{
			jobExecuteInfo.CancelFunc()
		}
	}
}

// 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulerPlan){
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting bool
	)

	// 调度和执行是2件事
	// 执行的任务可能运行很久，下次调度前还在运行，所以判断是否在运行中，如果还在运行中，则跳过本次执行
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name];jobExecuting{
		fmt.Println("还在执行")
		return
	}

	// 构建执行状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)

	// 保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	// TODO
	// G_executor.ExecuteJob(jobExecuteInfo)
}

// 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule()(schedulerAfter time.Duration){
	var (
		jobPlan *common.JobSchedulerPlan
		now time.Time
		nearTime *time.Time
	)

	// 如果任务表为空的话，随便睡眠多久
	if len(scheduler.jobPlanTable) == 0{
		schedulerAfter = 1 * time.Second
		return
	}

	// 获取当前时间
	now = time.Now()

	// 1. 遍历所有任务, 计算距离执行最少的时间
	for _, jobPlan = range scheduler.jobPlanTable{
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now){	// 如果任务下次执行时间过了当前时间，则尝试执行任务
			// TODO: 尝试执行任务
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 统计一个最近要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime){
			nearTime = &jobPlan.NextTime
		}
	}

	// 下次调度间隔(最近要执行的任务调度时间-当前时间)
	schedulerAfter = (*nearTime).Sub(now)
	return
}

// 处理任务结果
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult){
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

	// 生成执行日志
	if result.Err != common.ERR_LOCK_ALREADY_REQUIRED{
		jobLog := &common.JobLog{
			JobName: result.ExecuteInfo.Job.Name,
			Command: result.ExecuteInfo.Job.Command,
			Output: string(result.Output),
			PlanTime: result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime: result.StartTime.UnixNano() / 1000 / 1000,
			EndTime: result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil{
			jobLog.Err = result.Err.Error()
		}
		//TODO 把日志存储到mongodb
		G_LogSink.Append(jobLog)
	}
}

// 调度协程
func (scheduler *Scheduler) scheduleLoop(){
	// 定时任务 common.Job
	var (
		jobEvent *common.JobEvent
		schedulerAfter time.Duration
		schedulerTimer *time.Timer
		jobResult *common.JobExecuteResult
	)

	// 初始化任务调度睡眠间隔
	schedulerAfter = scheduler.TrySchedule()

	// 调度的延迟定时器
	schedulerTimer = time.NewTimer(schedulerAfter)

	for {
		select{
		case jobEvent = <- scheduler.jobEventChan:	// 监听任务变化事件
			scheduler.handleJobEvent(jobEvent)		// 对内存中维护的任务列表做增删改查
		case <- schedulerTimer.C:		// 休眠时间
		case jobResult = <- scheduler.jobResultChan:	// 接收执行完成任务结果，删除执行队列中任务，并保存任务结果
			scheduler.handleJobResult(jobResult)
		}
		// 重新计算调度时间
		schedulerAfter = scheduler.TrySchedule()
		schedulerTimer.Reset(schedulerAfter)
	}
}

// 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent){
	scheduler.jobEventChan <- jobEvent
}

// 回传任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult){
	scheduler.jobResultChan <- jobResult
}

// 初始化调度器
func InitScheduler(){
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulerPlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan: make(chan *common.JobExecuteResult, 1000),
	}

	// 启动调度协程
	go G_scheduler.scheduleLoop()
	return
}