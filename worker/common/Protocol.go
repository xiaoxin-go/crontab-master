package common

import (
	"context"
	"time"
	"github.com/gorhill/cronexpr"
)

// 定时任务
type Job struct{
	Name string	`json:"name"`	// 任务名称
	Command string `json:"command"`		// 任务命令
	CronExpr string	`json:"cronExpr"`	// 任务时间
}

type JobLog struct{
	JobName string `bson:"jobName"`
	Command string `bson:"command"`
	Err string `bson:"err"`
	Output string `bson:"output"`
	PlanTime int64 `bson:"planTime"`
	ScheduleTime int64 `bson:"scheduleTime"`
	StartTime int64 `bson:"startTime"`
	EndTime int64 `bson:"endTime"`
}

type LogBatch struct{
	Logs []interface{}
	StartTime time.Time
}

type JobEvent struct{
	EventType int
	Job *Job
}

func BuildJobEvent(eventType int, job *Job)(jobEvent *JobEvent){
	return &JobEvent{
		EventType: eventType,
		Job: job,
	}
}

type JobSchedulerPlan struct{
	Job *Job
	Expr *cronexpr.Expression
	NextTime time.Time
}

type JobExecuteInfo struct{
	Job *Job
	PlanTime time.Time		// 理论调度时间
	RealTime time.Time		// 现实调度时间
	CancelCtx context.Context
	CancelFunc context.CancelFunc
}

type JobExecuteResult struct{
	ExecuteInfo *JobExecuteInfo
	Output []byte
	Err error
	StartTime time.Time
	EndTime time.Time
}

func BuildJobSchedulePlan(job *Job)(jobSchedulerPlan *JobSchedulerPlan, err error){
	var (
		expr *cronexpr.Expression
	)

	if expr, err = cronexpr.Parse(job.CronExpr); err != nil{
		return
	}

	jobSchedulerPlan = &JobSchedulerPlan{
		Job: job,
		Expr: expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

func BuildJobExecuteInfo(plan *JobSchedulerPlan)(jobExecuteInfo *JobExecuteInfo){
	jobExecuteInfo = &JobExecuteInfo{
		Job: plan.Job,
		PlanTime: plan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}