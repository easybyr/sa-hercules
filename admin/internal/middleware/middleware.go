package middleware

import (
	"net/http"
	"strings"

	"github.com/example/sa-hercules/admin/internal/service"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/example/sa-hercules/admin/pkg/token"
	"github.com/gogf/gf/v2/net/ghttp"
)

const claimsKey = "auth.claims"

// Middleware 提供HTTP中间件。
type Middleware struct {
	tokens  *token.Manager
	service *service.Service
}

// New 创建中间件。
func New(tokenManager *token.Manager, svc *service.Service) *Middleware {
	return &Middleware{tokens: tokenManager, service: svc}
}

// CORS 允许本地前端调用API。
func (m *Middleware) CORS(r *ghttp.Request) {
	header := r.Response.Header()
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	if r.Method == http.MethodOptions {
		r.Response.Status = http.StatusNoContent
		r.Exit()
		return
	}
	r.Middleware.Next()
}

// Auth 校验Bearer Token和当前用户状态。
func (m *Middleware) Auth(r *ghttp.Request) {
	header := r.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		response.Error(r, apperror.New(http.StatusUnauthorized, "UNAUTHORIZED", "请先登录"))
		return
	}
	claims, err := m.tokens.Parse(parts[1])
	if err != nil {
		response.Error(r, apperror.New(http.StatusUnauthorized, "INVALID_TOKEN", "登录凭证无效或已过期"))
		return
	}
	if err = m.service.ValidateClaims(r.Context(), claims); err != nil {
		response.Error(r, err)
		return
	}
	r.SetCtxVar(claimsKey, claims)
	r.Middleware.Next()
}

// Claims 获取当前请求令牌声明。
func Claims(r *ghttp.Request) *token.Claims {
	value := r.GetCtxVar(claimsKey)
	if value.IsNil() {
		return nil
	}
	claims, _ := value.Val().(*token.Claims)
	return claims
}
