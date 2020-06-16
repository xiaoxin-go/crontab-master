package worker

import (
	"context"
	"crontab/worker/common"
	"crontab/worker/utils"
	"go.etcd.io/etcd/clientv3"
)

// 分布式锁
type JobLock struct{
	isLocked bool
	leaseId clientv3.LeaseID
	jobName string
	cancelFunc context.CancelFunc
}

func InitJobLock(jobName string)(jobLock *JobLock){
	jobLock = &JobLock{
		jobName: jobName,
	}
	return
}

// 尝试上锁
func (jobLock *JobLock) TryLock()(err error){
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
		leaseId clientv3.LeaseID
		keepRespChan <- chan *clientv3.LeaseKeepAliveResponse
		txn clientv3.Txn
		txnResp *clientv3.TxnResponse
		lockKey string
	)
	// 1. 创建租约(5秒)
	lease := utils.G_EtcdClient.Lease()
	if leaseGrantResp, err = utils.G_EtcdClient.GetLeaseGrant(context.TODO(), 5); err != nil{
		return
	}

	// context用于取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	leaseId = leaseGrantResp.ID

	// 2. 自动续租
	if keepRespChan, err = utils.G_EtcdClient.KeepAlive(cancelCtx, leaseId); err != nil{
		goto FAIL
	}
	// 处理自动续租应答的协程
	go func(){
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <- keepRespChan:
				if keepResp == nil{
					break
				}
			}
		}
	}()

	// 3. 创建事务txn
	txn = utils.G_EtcdClient.Txn(context.TODO())

	// 锁名称
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName

	// 4. 给事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	// 提交事务
	txnResp, err = txn.Commit()
	if err != nil{
		goto FAIL
	}

	// 5. 成功返回，失败释放租约
	if !txnResp.Succeeded{
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}

	// 抢锁成功
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	FAIL:
		cancelFunc()
		lease.Revoke(context.TODO(), leaseId)
		return
}

// 释放锁
func (jobLock *JobLock)Unlock(){
	if jobLock.isLocked{
		jobLock.cancelFunc()
		utils.G_EtcdClient.Lease().Revoke(context.TODO(), jobLock.leaseId)
	}
}
