package common

// 任务格式
type Job struct{
	Name string `json:"name"`
	Command string `json:"command"`
	Expr string `json:"expr"`
}

// 日志格式
type JobLog struct{
	JobName string `bson:"jobName"`		// 任务名
	Command string `bson:"command"`		// 脚本命令
	Err string `bson:"err"`				// 错误原因
	Output string `bson:"output"`		// 脚本输出
	PlanTime int64 `bson:"planTime"`	// 计划执行时间
	ScheduleTime int64 `bson:"scheduleTime"`	// 实际调度时间
	StartTime int64 `bson:"startTime"`		// 任务执行时间
	EndTime int64 `bson:"endTime"`		// 任务结束时间
}

// 任务日志过滤条件
type JobLogFilter struct{
	JobName string `bson:"jobName"`
}

// 任务日志排序条件
type SortLogByStartTime struct{
	SortOrder int `bson:"startTime"`
}
