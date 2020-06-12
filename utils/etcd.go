package utils

import (
	"context"
	"crontab/conf"
	"encoding/json"
	"errors"
	"github.com/astaxie/beego"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type EtcdClient struct{
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}

// 获取租约
func (this *EtcdClient) GetLease(ttl int64)(*clientv3.LeaseGrantResponse, error){
	return this.lease.Grant(context.TODO(), ttl)
}

// 写入数据，带租约的写入
func (this *EtcdClient) PutLease(key string, value interface{}, ttl int64)(result []byte, err error){
	leaseGrantResponse, err := this.GetLease(ttl)
	if err != nil{
		return
	}
	// 将保存数据对象，转换成json字符串存入etcd
	return this.Put(key, value, clientv3.WithLease(leaseGrantResponse.ID))
}

// 获取etcd中单条数据
func (this *EtcdClient) Get(key string)(result []byte, err error){
	getResp, err := this.kv.Get(context.TODO(), key)
	if err != nil{
		return
	}
	if getResp.Count == 0{
		err = errors.New("data is empty")
	}else{
		result = getResp.Kvs[0].Value
	}
	return
}

// 从etcd中获取多条数据
func (this *EtcdClient) GetList(key string)(result map[string][]byte, err error){
	// 查询etcd
	getResp, err := this.kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil{
		return
	}
	// 将数据返回
	result = make(map[string][]byte)
	for _, kvPair := range getResp.Kvs{
		result[string(kvPair.Key)] = kvPair.Value
	}
	return
}

// 向etcd中添加数据
func (this *EtcdClient) Put(key string, value interface{}, opts ...clientv3.OpOption)(result []byte, err error){
	// 将保存数据对象，转换成json字符串存入etcd
	byteValue, err := json.Marshal(value)
	if err != nil{
		return
	}
	putResp, err := this.kv.Put(context.TODO(), key, string(byteValue), opts...)
	if err != nil{
		return
	}

	// 返回旧数据
	result = putResp.PrevKv.Value
	return
}

// 从etcd删除数据
func (this *EtcdClient) Delete(key string)(result []byte, err error){
	// 从etcd中删除key
	delResp, err := this.kv.Delete(context.TODO(), key)
	if err != nil{
		return
	}
	// 如果返回数据大于0，取删除的第一条数据
	if len(delResp.PrevKvs) != 0{
		result = delResp.PrevKvs[0].Value
	}
	return
}

var (
	G_EtcdClient *EtcdClient
)

func init(){
	// 初始化配置
	config := clientv3.Config{
		Endpoints: conf.EtcdEndpoints,
		DialTimeout: time.Duration(conf.EtcdDialTimeout) * time.Millisecond,
	}

	// 连接etcd
	client, err := clientv3.New(config)
	if err != nil{
		beego.Info("etcd连接异常： ", err)
		panic(err)
	}

	// 得到kv和lease
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	// 赋值全局单例
	G_EtcdClient = &EtcdClient{
		client: client,
		kv: kv,
		lease: lease,
	}
	return
}