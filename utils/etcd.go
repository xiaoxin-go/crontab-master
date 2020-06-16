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
	watcher clientv3.Watcher
}

func (this *EtcdClient)Txn(ctx context.Context)clientv3.Txn{
	return this.kv.Txn(ctx)
}

func (this *EtcdClient)Watcher()clientv3.Watcher{
	return this.watcher
}

func (this *EtcdClient)Lease()clientv3.Lease{
	return this.lease
}

// 获取授权租约
func (this *EtcdClient) GetLeaseGrant(ctx context.Context, ttl int64)(*clientv3.LeaseGrantResponse, error){
	return this.lease.Grant(ctx, ttl)
}

// 续租
func (this *EtcdClient) KeepAlive(ctx context.Context, id clientv3.LeaseID)(<-chan *clientv3.LeaseKeepAliveResponse, error){
	return this.lease.KeepAlive(ctx, id)
}

// 写入数据，带租约的写入
func (this *EtcdClient) PutLease(ctx context.Context, key string, value interface{}, ttl int64)(result []byte, err error){
	leaseGrantResponse, err := this.GetLeaseGrant(context.TODO(), ttl)
	if err != nil{
		return
	}
	// 将保存数据对象，转换成json字符串存入etcd
	return this.Put(ctx, key, value, clientv3.WithLease(leaseGrantResponse.ID))
}

func (this *EtcdClient)Get(ctx context.Context, key string, opts ...clientv3.OpOption)(*clientv3.GetResponse, error){
	return this.kv.Get(ctx, key, opts...)
}

// 获取etcd中单条数据
func (this *EtcdClient) GetOne(ctx context.Context, key string, opts ...clientv3.OpOption)(result []byte, err error){
	getResp, err := this.Get(ctx, key, opts...)
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

// 解析从etcd中获取的数据
func (this *EtcdClient) ReadValueList(response *clientv3.GetResponse, result interface{})(err error){
	valueList := make([]string, 0)
	for _, kvPair := range response.Kvs{
		valueList = append(valueList, string(kvPair.Value))
	}
	bytes, err := json.Marshal(valueList)
	if err != nil{
		return
	}
	err = json.Unmarshal(bytes, result)
	return
}

// 解析从Etcd中获取的key
func (this *EtcdClient) ReadKeyList(response *clientv3.GetResponse)(result []string, err error){
	result = make([]string, 0)
	for _, kvPair := range response.Kvs{
		result = append(result, string(kvPair.Key))
	}
	return
}

// 向etcd中添加数据
func (this *EtcdClient) Put(ctx context.Context, key string, value interface{}, opts ...clientv3.OpOption)(result []byte, err error){
	// 将保存数据对象，转换成json字符串存入etcd
	byteValue, err := json.Marshal(value)
	if err != nil{
		return
	}
	putResp, err := this.kv.Put(ctx, key, string(byteValue), opts...)
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
	watcher := clientv3.NewWatcher(client)

	// 赋值全局单例
	G_EtcdClient = &EtcdClient{
		client: client,
		kv: kv,
		lease: lease,
		watcher: watcher,
	}
	return
}