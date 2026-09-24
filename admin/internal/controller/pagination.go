package controller

import (
	"github.com/example/sa-hercules/admin/internal/model"
	"github.com/gogf/gf/v2/net/ghttp"
)

func pageRequest(r *ghttp.Request) model.PageRequest {
	return model.NewPageRequest(r.Get("page").Int(), r.Get("page_size").Int())
}
