package main

import (
	_ "crontab/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}

