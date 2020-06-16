package worker

import (
	"context"
	"crontab/worker/common"
	"crontab/worker/utils"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

type Register struct{
	localIP string
}

// 注册到/cron/workoers/IP, 并自动续租
func (register *Register) keepOnline(){
	var (
		keepAliveChan <- chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
	)
	for {
		// 注册key
		regKey := common.Job_WORKER_DIR + register.localIP

		// 创建租约
		leaseGrantResp, err := utils.G_EtcdClient.GetLeaseGrant(context.TODO(), 10)
		if err != nil{
			goto RETRY
		}

		// 自动续租
		if keepAliveChan, err = utils.G_EtcdClient.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil{
			goto RETRY
		}

		// 创建取消上下文
		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// 注册到ETCD
		if _, err = utils.G_EtcdClient.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil{
			goto RETRY
		}

		for{
			select{
			case keepAliveResp = <- keepAliveChan:
				if keepAliveResp == nil{ // 续租失败
					goto RETRY
				}
			}
		}

		// 续租发生异常，休息一秒，重新续租
		RETRY:
			time.Sleep(1 * time.Second)
			if cancelFunc != nil{
				cancelFunc()
			}

	}

}

// 获取本地IP
func getLocalIP()(ipv4 string, err error){
	// 获取所有网卡
	addrList, err := net.InterfaceAddrs()
	if err != nil{
		return
	}
	 for _, addr := range addrList{
		// ipv4, ipv6, 需要对IP地址做反解
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback(){
			// 这个网络地址是IP地址：ipv4或ipv6,跳过ipv6
			if ipNet.IP.To4() != nil{
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

// 服务注册
func InitRegister()(err error){
	// 获取本地IP
	localIp, err := getLocalIP()
	if err != nil{
		return
	}
	// 注册服务, 不停向etcd续租，每10秒续租一次，若进程挂掉，则节点消失
	register := Register{localIP: localIp}
	go register.keepOnline()
	return
}
