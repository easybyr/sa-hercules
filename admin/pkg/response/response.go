package response

import (
	"errors"
	"net/http"

	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Response 是统一接口响应。
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 写入成功响应。
func OK(r *ghttp.Request, data any) {
	r.Response.WriteJsonExit(Response{Code: 0, Message: "success", Data: data})
}

// Created 写入创建成功响应。
func Created(r *ghttp.Request, data any) {
	r.Response.Status = http.StatusCreated
	r.Response.WriteJsonExit(Response{Code: 0, Message: "created", Data: data})
}

// Error 写入失败响应。
func Error(r *ghttp.Request, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		r.Response.Status = appErr.Status
		r.Response.WriteJsonExit(Response{Code: appErr.Status, Message: appErr.Message})
		return
	}
	r.Response.Status = http.StatusInternalServerError
	r.Response.WriteJsonExit(Response{Code: http.StatusInternalServerError, Message: "服务暂时不可用"})
}
