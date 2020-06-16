package worker

import (
	"context"
	"crontab/worker/common"
	"crontab/worker/utils"
	"fmt"
	"time"
)
var(
	G_LogSink *LogSink
)

// mongodb存储日志
type LogSink struct{
	logChan chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

// 日志存储协程
func (logSink *LogSink)writeLoop(){
	var (
		logBatch *common.LogBatch
	)
	for {
		select{
		case log := <- logSink.logChan:
			// 把这条log写入到mongoddb中
			// 每次插入需要等待mongodb的一次请求往返，耗时可能因为网络慢花费较长的时间
			// 把日志存到一个列表中，当达到一定数量或者一定时插入日志
			if logBatch == nil{
				logBatch = &common.LogBatch{
					// 开始时间，超时5秒，和之后时间对比，若超过此时间，则执行插入操作
					StartTime: time.Now().Add(time.Duration(G_Config.JobLogCommitTimeout) * time.Millisecond),
				}
			}

			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			// 如果批次满了，或者达到一定时间，则存储日志
			if len(logBatch.Logs) >= G_Config.JobLogBatchSize || logBatch.StartTime.Before(time.Now()){
				logSink.saveLogs(logBatch)

				// 清空logBatch
				logBatch = nil
			}
		}
	}
}

// 保存日志，不管是否发生异常
func (logSink *LogSink)saveLogs(logBatch *common.LogBatch){
	_, err := utils.G_mongoClient.InsertMany(context.TODO(), logBatch.Logs)
	if err != nil{
		fmt.Println("插入日志异常： ", err.Error())
	}
}

// 发送日志，将日志写入到队列
func (logSink *LogSink) Append(jobLog *common.JobLog){
	select{
	case logSink.logChan <- jobLog:
	default:
	}
}

func InitLogSink()(err error){
	G_LogSink = &LogSink{
		logChan: make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch),
	}
	go G_LogSink.writeLoop()
	return
}
