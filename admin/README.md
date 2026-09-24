# Hercules Admin Backend

后端使用 GoFrame v2、MySQL 8、JWT 和 bcrypt，实现组织隔离的用户、角色与权限管理。

## 目录说明

- `cmd/server`：服务入口与依赖组装。
- `config`：默认、开发、测试、生产四套 GoFrame 配置。
- `internal/controller`：HTTP 路由和参数处理。
- `internal/middleware`：CORS、JWT 和有效用户校验。
- `internal/service`：组织隔离、RBAC、CRUD 与事务业务规则。
- `internal/repository`：所有数据库访问。
- `internal/bootstrap`：幂等初始化超级管理员和内置权限。
- `internal/consts`：状态等固定值枚举。
- `pkg`：密码、令牌、统一响应、业务错误和业务 ID 工具。
- `resource/migrations`：MySQL 初始化脚本。

## 启动

先在项目根目录启动 MySQL，再运行：

```bash
go mod download
go run ./cmd/server
```

未设置 `APP_ENV` 时默认读取 `config/config.dev.yaml`。可选值为 `default`、`dev`、`test`、`prod`，其中 `default` 对应 `config/config.yaml`。默认端口 `8000`，Swagger UI 位于 `/swagger/`。

```bash
APP_ENV=dev go run ./cmd/server
APP_ENV=test go run ./cmd/server
APP_ENV=prod JWT_SECRET='replace-me' go run ./cmd/server
```

Controller 采用“资源 + API 版本 + 动作”的文件命名，例如 `user_v1_create_org_member.go` 实现 `CreateOrgMember`。Service、Repository 与 Model 也按组织、用户、权限、角色、访问控制等职责拆分。

## 创建规则

- `POST /api/organizations` 不接收 `org_id`，后端生成 8 位组织 ID；`name` 必填，`parent_id`
  保存直接上级组织主键且根组织为 `null`。
- 注册和 `POST /api/users` 不接收 `uid`，后端生成全局唯一的 10 位 UID。管理员创建用户时密码可
  为空，此时生成 16 位随机初始密码，并只在创建响应中返回一次。
- `POST /api/roles` 不接收 `role_id`，后端生成全局唯一角色 ID，并校验组织内角色名称唯一。
- 权限编码必须采用 `report.read` 形式的小写点分格式；权限编码和名称都在组织内唯一。

## 测试

```bash
go test ./...
```

如果本机 Go 缓存目录权限受限，可临时指定缓存：

```bash
GOCACHE=/tmp/hercules-go-cache go test ./...
```

## 权限模型

访问令牌包含 `uid` 和当前 `org_id`。每个受保护请求会再次检查用户与组织状态。权限查询只连接同一 `org_id` 下的用户角色、角色权限和权限定义，因此父子组织不会自动继承权限。超级管理员 `uid=1000000000` 是唯一跨组织例外。
