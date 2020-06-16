master

接口： 
新建任务： /job/save
查看任务： /job/list
删除任务： /job/delete
杀死任务： /job/kill
查看日志： /job/log

节点健康： /worker

相关连接： etcd(任务增删改查) mongo(日志查看)
配置管理

worker
scheduler	调度模块
job			任务管理
log			日志管理
register	服务注册
