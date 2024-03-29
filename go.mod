module crontab

go 1.14

require (
	github.com/astaxie/beego v1.12.1
	github.com/coreos/etcd v3.3.22+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	go.etcd.io/etcd v3.3.22+incompatible
	go.mongodb.org/mongo-driver v1.3.4
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/grpc v1.29.1 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
