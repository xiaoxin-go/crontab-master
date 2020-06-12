package utils

import (
	"crontab/common"
	"fmt"
	"github.com/astaxie/beego"
)

type HttpController struct{
	beego.Controller
}

func (h *HttpController) response(){
	h.ServeJSON()
	h.StopRun()
}

// 成功返回
func (h *HttpController) HttpSuccess(respData interface{}, message string){
	h.Data["json"] = common.Success(respData, message)
	h.response()
}

// 异常返回
func (h *HttpController) HttpError(err error, message string, response func(string) *common.HttpResponse){
	if err != nil{
		h.Data["json"] = response(fmt.Sprintf("Message: %s, Error: %s", message, err.Error()))
		h.response()
	}
}

// 参数异常返回
func (h *HttpController) HttpParamsError(err error, message string){
	h.HttpError(err, message, common.ParamsError)
}

// 服务器异常返回
func (h *HttpController) HttpServerError(err error, message string){
	h.HttpServerError(err, message)
}

