package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/example/sa-hercules/admin/internal/bootstrap"
	appconfig "github.com/example/sa-hercules/admin/internal/config"
	"github.com/example/sa-hercules/admin/internal/controller"
	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/internal/repository"
	"github.com/example/sa-hercules/admin/internal/service"
	"github.com/example/sa-hercules/admin/pkg/apperror"
	"github.com/example/sa-hercules/admin/pkg/response"
	"github.com/example/sa-hercules/admin/pkg/token"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.GetInitCtx()
	environment, err := appconfig.Load()
	if err != nil {
		g.Log().Fatal(ctx, "加载配置失败:", err)
	}
	g.Log().Info(ctx, "当前运行环境:", environment)
	store := repository.New()
	secret := g.Cfg().MustGet(ctx, "security.jwtSecret").String()
	if environmentSecret := os.Getenv("JWT_SECRET"); environmentSecret != "" {
		secret = environmentSecret
	}
	expireHours := g.Cfg().MustGet(ctx, "security.jwtExpireHours", 24).Int()
	tokenManager := token.NewManager(secret, time.Duration(expireHours)*time.Hour)
	svc := service.New(store, tokenManager)
	mw := middleware.New(tokenManager, svc)
	api := controller.New(svc)
	server := g.Server()
	ready := make(chan struct{})
	server.Use(func(r *ghttp.Request) {
		select {
		case <-ready:
			r.Middleware.Next()
		default:
			response.Error(r, apperror.New(503, "SERVICE_INITIALIZING", "服务正在初始化"))
		}
	})
	api.RegisterRoutes(server, mw)
	if err := server.Start(); err != nil {
		g.Log().Fatal(ctx, "启动 API 路由失败:", err)
	}
	adminUsername := strings.TrimSpace(os.Getenv("SUPER_ADMIN_USERNAME"))
	adminPassword := os.Getenv("SUPER_ADMIN_PASSWORD")
	if adminUsername == "" || adminPassword == "" {
		g.Log().Fatal(ctx, "SUPER_ADMIN_USERNAME 和 SUPER_ADMIN_PASSWORD 必须配置")
	}
	if err := bootstrap.EnsureSuperAdmin(context.Background(), store, server.GetRoutes(), bootstrap.SuperAdminConfig{
		Username: adminUsername,
		Password: adminPassword,
	}); err != nil {
		g.Log().Fatal(ctx, "初始化超级管理员失败:", err)
	}
	close(ready)
	ghttp.Wait()
}
