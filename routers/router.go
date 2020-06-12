package routers

import (
	"crontab/controllers"
	"github.com/astaxie/beego"
)

func init() {
    //beego.Router("/", &controllers.MainController{})
    beego.Router("/job", &controllers.JobController{}, "get:GetList;post:Save")
    beego.Router("/job/:name", &controllers.JobController{}, "post:Kill;delete:Del")
    beego.Router("/job/:name/log", &controllers.LogController{}, "get:GetList")
    beego.Router("/worker", &controllers.WorkController{}, "get:GetList")
}
