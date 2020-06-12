package conf

import "github.com/astaxie/beego"

var (
	EtcdEndpoints []string
	EtcdDialTimeout int64
	MongodbUri string
)


func init(){
	EtcdEndpoints = beego.AppConfig.DefaultStrings("etcdendpoints", []string{"127.0.0.1:2379"})
	EtcdDialTimeout = beego.AppConfig.DefaultInt64("etcddialtimeout", 5000)
	beego.Info(EtcdEndpoints)
	MongodbUri = beego.AppConfig.String("mongodburi")
}