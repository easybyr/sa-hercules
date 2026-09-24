package main

import (
	"context"
	"os"
	"time"

	"github.com/example/sa-hercules/admin/internal/bootstrap"
	appconfig "github.com/example/sa-hercules/admin/internal/config"
	"github.com/example/sa-hercules/admin/internal/controller"
	"github.com/example/sa-hercules/admin/internal/middleware"
	"github.com/example/sa-hercules/admin/internal/repository"
	"github.com/example/sa-hercules/admin/internal/service"
	"github.com/example/sa-hercules/admin/pkg/token"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
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
	if err := bootstrap.EnsureSuperAdmin(context.Background(), store); err != nil {
		g.Log().Fatal(ctx, "初始化超级管理员失败:", err)
	}
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
	api.RegisterRoutes(server, mw)
	server.Run()
}
