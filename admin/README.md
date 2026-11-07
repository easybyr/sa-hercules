# Hercules Admin Backend

后端使用 GoFrame v2、MySQL 8、JWT 和 bcrypt，实现全局用户与租户隔离的角色和权限管理。

## 目录说明

- `cmd/server`：服务入口与依赖组装。
- `config`：默认、开发、测试、生产四套 GoFrame 配置。
- `internal/controller`：HTTP 路由和参数处理。
- `internal/middleware`：CORS、JWT 和有效用户校验。
- `internal/service`：租户隔离、RBAC、CRUD 与事务业务规则。
- `internal/repository`：所有数据库访问。
- `internal/bootstrap`：幂等初始化超级管理员、API 登记和内置权限。
- `internal/consts`：状态等固定值枚举。
- `pkg`：密码、令牌、统一响应、业务错误和 ID 工具。
- `resource/migrations/004_init.sql`：当前唯一需要的 MySQL 初始化脚本。

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
APP_ENV=prod JWT_SECRET='replace-me' SUPER_ADMIN_USERNAME='admin' SUPER_ADMIN_PASSWORD='replace-me' go run ./cmd/server
```

## 数据与接口约定

- 执行 `resource/migrations/004_init.sql` 初始化 `hercules_admin`。migrations 目录中其他 SQL 已废弃。
- 首次启动会使用 `SUPER_ADMIN_USERNAME` 和 `SUPER_ADMIN_PASSWORD` 创建超级管理员；Docker Compose 从项目根目录 `.env` 读取它们。两个变量均为必填，已有超级管理员不会被覆盖。
- 租户保存在 `sys_tenant`；`id` 是内部主键，`code` 是可用于登录和展示的业务编码，`parent_code` 指向上级租户编码。
- 新建租户会自动创建 `tenant_admin`（租户管理员）系统角色，并同步该租户全部权限。系统角色不可单独修改或删除；超级管理员删除租户时会一并清理该租户的成员关系、角色、权限定义及失效的 API 权限绑定。
- 租户的管理列表和树仅超级管理员可访问；普通租户用户只能在当前租户范围内进行已获授权的权限管理。
- 用户保存在全局 `sys_user`，用户名全局唯一；`sys_tenant_member` 维护用户与租户的关系。
- 登录和注册使用 `tenant_code`。访问令牌包含 `uid` 与当前 `tenant_id`（`sys_tenant.id`）。
- `POST /api/tenants` 自动生成 8 位租户编码；`POST /api/users` 自动生成 UID。管理员创建用户时密码可为空，此时生成一次性初始密码并只在创建响应中返回。
- 角色编码在租户内唯一，角色记录以 `sys_role.id` 作为 API 的 `role_id`。角色权限通过 `sys_role_permission.permission_id` 关联权限主键；成员角色绑定到 `sys_tenant_member.id`。
- 权限节点类型使用 `MENU`、`BUTTON`、`API`、`DATA` 字符串，编码前缀分别为 `menu.`、`button.`、`api.`、`data.`。按钮和 API 节点必须关联已登记 API；数据节点可选关联。
- `sys_api_permission` 是全局 API 路由目录，保存 `service_name`、`http_method`、`api_path`、`path_type` 与 `permission_code`。启动时会幂等登记受控路由。
- `sys_permission_oper_log` 记录租户内权限、角色和用户授权变更。

## 权限检查

`GET /api/access/user-permission?tenant_id=1&uid=1000000000&permission_code=menu.user` 检查指定用户是否属于指定租户并具备权限编码。接口要求登录，成功响应的 `data` 是布尔值。
