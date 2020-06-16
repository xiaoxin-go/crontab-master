package worker

import (
	"crontab/worker/common"
	"math/rand"
	"os/exec"
	"time"
)

type Executor struct{

}

var (
	G_executor *Executor
)

// 执行一个任务
func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo){
	go func(){
		var (
			cmd *exec.Cmd
			output []byte
			err error
			result *common.JobExecuteResult
			jobLock *JobLock
		)

		// 任务结果
		result = &common.JobExecuteResult{
			ExecuteInfo:info,
			Output: make([]byte, 0),
		}

		// 获取分布式锁，抢到锁则执行任务
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)

		// 初始化任务开始时间
		result.StartTime = time.Now()

		// 随便睡眠，避免一直抢占锁
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))

		if err = jobLock.TryLock(); err != nil{		// 未抢到锁
			result.Err = err
			result.EndTime = time.Now()
		}else{
			result.StartTime = time.Now()
			cmd = exec.CommandContext(info.CancelCtx, "c:\\cygwin64\\bin\\bash.exe", "-c", info.Job.Command)

			// 执行并捕获输出
			output, err = cmd.CombinedOutput()

			// 记录任务结束时间
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err

			// 任务执行完成后，释放锁
			jobLock.Unlock()
		}
		// 任务执行完成后，把任务结果推给调度器
		G_scheduler.PushJobResult(result)
	}()
}

// 初始化执行器
func InitExecutor(){
	G_executor = &Executor{}
	return
}
