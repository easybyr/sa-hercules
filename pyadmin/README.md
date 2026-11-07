# PyAdmin

PyAdmin 是 Hercules Admin 后端（`admin`，Go + GoFrame + MySQL）的 **Python3 + FastAPI + PostgreSQL** 重实现，
对外提供与 `admin-web` 完全兼容的 HTTP 接口。组织、用户、角色、权限、登录注册、个人中心与 RBAC 组织隔离
等全部能力均已迁移，`admin-web` 无需任何改动即可对接。

## 技术栈

| 层面 | 选型 |
| --- | --- |
| 语言 / 运行时 | Python 3.9+（镜像使用 3.12） |
| Web 框架 | FastAPI 0.115 + Uvicorn |
| 数据库 | PostgreSQL 16 + SQLAlchemy 2.0（asyncpg 异步驱动） |
| 认证 | JWT（HS256，PyJWT）+ bcrypt 密码哈希 |
| 配置 | YAML 分环境配置 + 环境变量覆盖 |
| 测试 | pytest + httpx（SQLite 端到端）、可选真实 PostgreSQL 集成测试 |

## 目录结构

```text
pyadmin
├── app
│   ├── config.py              # 分环境配置加载与环境变量覆盖
│   ├── main.py                # 应用装配：生命周期、CORS、统一异常、路由
│   ├── __main__.py            # python -m app 启动入口
│   ├── consts                 # 枚举与常量（状态、业务 ID、内置权限、正则、分页）
│   ├── db                     # 异步引擎与会话工厂
│   ├── model                  # SQLAlchemy ORM 实体
│   ├── schema                 # 请求/响应模型与字段校验
│   ├── repository             # 只做数据库读写，不含业务逻辑
│   ├── service                # 业务编排：组织隔离、乐观锁、事务边界
│   ├── controller             # 只接收请求、调用 service、返回统一响应
│   ├── middleware             # 认证依赖（Bearer 解析、登录态与权限校验）
│   └── pkg                    # 通用能力：错误、响应、ID、密码、令牌、时间、校验提示
├── config                     # config.yaml / dev / test / prod
├── resources/migrations       # PostgreSQL 初始化 DDL
├── deploy/nginx.conf          # admin-web 反向代理到 pyadmin
├── tests                      # 单元测试 + 端到端接口测试 + Postgres 集成测试
├── docker-compose.yaml        # PostgreSQL + pyadmin + admin-web
├── Dockerfile
├── requirements.txt
└── requirements-dev.txt
```

## 与 Go 版本的分层对应

| Go `admin` | PyAdmin | 说明 |
| --- | --- | --- |
| `internal/controller/*.go` | `app/controller/*_v1.py` | 一资源一模块，与 `user_v1_*` 命名同源；只解析入参、鉴权、返回响应 |
| `internal/service/*.go` | `app/service/*.py` | 业务规则；`AccessGuard` 统一实现权限与组织隔离 |
| `internal/repository/*.go` | `app/repository/*.py` | 纯数据库访问；`Store` 由各 repository mixin 组合 |
| `internal/model` | `app/model`（ORM）+ `app/schema`（接口模型） | 实体与接口视图分离 |
| `internal/middleware` | `app/middleware/auth.py` | FastAPI 依赖式认证 |
| `pkg/*` | `app/pkg/*` | apperror / response / idgen / password / token |
| `internal/bootstrap` | `app/service/bootstrap.py` | 幂等初始化超级管理员与内置权限 |
| `resource/migrations` | `resources/migrations` | MySQL DDL 等价改写为 PostgreSQL |

设计约束与 Go 版本一致：`controller` 不写业务逻辑、`service` 不写 SQL、`repository` 不写业务；
统一响应体 `{code, message, data}`；`version` 乐观锁；数据库不使用外键与 CHECK 约束。

## 接口清单

除健康检查、登录、注册与组织精确搜索外，均需 `Authorization: Bearer <token>`；表格中的“权限”指入口权限编码。

| 方法 | 路径 | 权限 | 说明 |
| --- | --- | --- | --- |
| GET | `/api/health` | 无 | 健康检查，含一次数据库探活 |
| POST | `/api/auth/login` | 无 | 登录，返回 JWT 与用户信息 |
| POST | `/api/auth/register` | 无 | 在有效组织内注册（201） |
| GET | `/api/public/organizations/search?q=` | 无 | 注册页按完整名称/组织 ID 精确搜索 |
| GET | `/api/me` | 登录 | 当前用户、组织、角色、权限与按角色分组的权限 |
| PUT | `/api/me` | 登录 | 修改自己的资料与密码 |
| GET | `/api/access/check?permission=&role_id=&org_id=` | 登录 | 权限、角色、组织三类访问校验 |
| GET | `/api/organizations` | organization.read | 分页 + 关键词查询可见组织 |
| GET | `/api/organizations/tree` | organization.read | 组织树（不分页，可 `q` 过滤） |
| POST | `/api/organizations` | organization.create | 新建组织（201，仅超级管理员业务上可成功） |
| PUT | `/api/organizations/{id}` | organization.update | 修改组织（`version` 乐观锁） |
| DELETE | `/api/organizations/{id}` | organization.delete | 删除空组织 |
| GET | `/api/users` | user.read | 分页 + 关键词 + `org_id` 查询用户 |
| POST | `/api/users` | user.create | 新建用户（201，密码留空返回一次性初始密码） |
| GET | `/api/users/{uid}` | user.read | 用户详情（角色、权限、直接权限编码、组织） |
| PUT | `/api/users/{uid}` | user.update | 修改用户 |
| DELETE | `/api/users/{uid}` | user.delete | 删除用户并清理关联 |
| PUT | `/api/users/{uid}/roles` | user.assign_role | 替换用户角色（同时清理直接权限） |
| PUT | `/api/users/{uid}/permissions` | user.assign_permission | 分配用户直接权限 |
| PUT | `/api/users/{uid}/organization` | user.assign_organization | 跨组织移动用户（仅超级管理员） |
| GET | `/api/permissions` | permission.read | 分页 + 关键词 + `org_id` 查询权限 |
| POST | `/api/permissions` | permission.create | 新建权限（201） |
| PUT | `/api/permissions/{org_id}/{code}` | permission.update | 修改权限 |
| DELETE | `/api/permissions/{org_id}/{code}` | permission.delete | 删除权限并清理角色关联 |
| GET | `/api/roles` | role.read | 分页 + 关键词 + `org_id` 查询角色 |
| POST | `/api/roles` | role.create | 新建角色（201，可同时提交权限编码） |
| GET | `/api/roles/{org_id}/{role_id}` | role.read | 角色详情与权限编码 |
| PUT | `/api/roles/{org_id}/{role_id}` | role.update | 修改角色（不传 `permission_codes` 时保留原权限） |
| DELETE | `/api/roles/{org_id}/{role_id}` | role.delete | 删除角色并清理关联 |
| PUT | `/api/roles/{org_id}/{role_id}/permissions` | role.assign_permission | 替换角色权限 |

交互式文档：`http://127.0.0.1:8000/docs`（Swagger UI）、`/redoc`、`/api.json`；
`/swagger` 会 307 跳转到 `/docs`，兼容原 Go 版本的入口习惯。

## 统一响应与错误码

```json
{ "code": 0, "message": "success", "data": {} }
```

- 成功：`code = 0`，`message = "success"`；创建接口返回 HTTP 201 且 `message = "created"`。
- 失败：`code` 等于 HTTP 状态码，`message` 为中文说明，`data` 为 `null`。
- 分页 `data` 固定为 `{list, total, page, page_size}`，默认 `page=1`、`page_size=10`，`page_size` 上限 1000；
  列表统一按 `updated_at DESC, id DESC` 排序。

| 状态 | 典型场景 | `message` 示例 |
| --- | --- | --- |
| 400 | 参数校验失败、状态非法、组织环路 | 用户名长度需在3到64个字符之间 / 状态值无效 |
| 401 | 未带 token、token 失效、用户被禁用后旧 token | 请先登录 / 登录凭证无效或已过期 / 登录状态已失效 |
| 403 | 缺权限、跨组织访问、非超级管理员管理组织 | 没有访问该资源的权限 / 不能访问其他组织的数据 |
| 404 | 资源不存在 | 组织不存在 |
| 409 | 唯一冲突、乐观锁冲突、组织非空 | 用户名已存在 / 数据已被其他请求修改，请刷新后重试 |
| 500 | 未预期异常（不泄露内部细节） | 服务暂时不可用 |

状态字段在数据库存 `int`（`1` 正常、`2` 禁用），接口统一返回 `"normal"` / `"disabled"`；
时间字段按 RFC3339 毫秒输出，如 `2026-11-06T23:17:52.123+08:00`。

## 数据库结构

初始化脚本：[`resources/migrations/001_init.sql`](./resources/migrations/001_init.sql)。所有表均包含
`id BIGSERIAL` 主键、业务标识、`extra JSONB`、`version INT`、`created_at/updated_at TIMESTAMPTZ(3)`；
不使用外键与 CHECK 约束，限定值由代码校验；表与字段均带 `COMMENT`。

### organization（组织表）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL | 主键 |
| org_id | BIGINT | 组织业务 ID，唯一，8 位数字 |
| name | VARCHAR(100) | 组织名称 |
| parent_id | BIGINT NULL | 直接上级组织的 `id`，根组织为空 |
| description | VARCHAR(500) | 描述 |
| status | INT | 1 正常 / 2 禁用 |
| extra / version / created_at / updated_at | JSONB / INT / TIMESTAMPTZ(3) / TIMESTAMPTZ(3) | 通用字段 |

索引：`uk_organization_org_id`、`idx_organization_parent_id`、`idx_organization_name`、`idx_organization_updated_at`。

### admin_user（用户表）

Go 版本表名为 `user`，在 PostgreSQL 中是保留字，因此改名为 `admin_user`（接口与前端完全无感知）。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGSERIAL | 主键 |
| uid | BIGINT | 全局唯一用户 ID，10 位数字 |
| username | VARCHAR(64) | 用户名，全局唯一 |
| nickname / phone_num / email | VARCHAR | 昵称、手机号、邮箱 |
| password | VARCHAR(255) | bcrypt 哈希 |
| org_id | BIGINT | 所属组织业务 ID |
| status | INT | 1 正常 / 2 禁用 |
| extra / version / created_at / updated_at | JSONB / INT / TIMESTAMPTZ(3) / TIMESTAMPTZ(3) | 通用字段 |

索引：`uk_admin_user_uid`、`uk_admin_user_username`、`idx_admin_user_org_id`、`idx_admin_user_org_updated_at`。

### permission（权限表）

`id`、`code`、`name`、`description`、`org_id`、`extra`、`version`、`created_at`、`updated_at`；
唯一索引 `uk_permission_org_code (org_id, code)` 与 `uk_permission_org_name (org_id, name)`。

### role（角色表）

`id`、`role_id`（全局唯一）、`name`、`description`、`org_id`、`extra`、`version`、`created_at`、`updated_at`；
唯一索引 `uk_role_role_id (role_id)` 与 `uk_role_org_name (org_id, name)`。
用户直接权限通过隐藏的个人角色实现，其 `role_id = -uid`（负数），列表接口默认隐藏。

### user_role（用户角色关联表）

`id`、`uid`、`role_id`、`extra`、`version`、`created_at`、`updated_at`；唯一索引 `uk_user_role_uid_role_id (uid, role_id)`。
关联表不额外增加业务 ID，`(uid, role_id)` 即业务唯一键（与 Go 版本一致）。

### role_permission（角色权限关联表）

`id`、`role_id`、`permission_code`、`extra`、`version`、`created_at`、`updated_at`；
唯一索引 `uk_role_permission_role_code (role_id, permission_code)`；`(role_id, permission_code)` 即业务唯一键。

### 与 MySQL 版本的差异（行为等价）

| 方面 | MySQL（原实现） | PostgreSQL（PyAdmin） |
| --- | --- | --- |
| 自增主键 | `AUTO_INCREMENT` | `BIGSERIAL` |
| 毫秒时间 | `DATETIME(3)` | `TIMESTAMPTZ(3)`，会话时区 `Asia/Shanghai` |
| `updated_at` 自动刷新 | `ON UPDATE CURRENT_TIMESTAMP(3)` | `pyadmin_set_updated_at()` 触发器 + ORM `onupdate` |
| 大小写不敏感匹配 | `utf8mb4_0900_ai_ci` 排序规则 | 查询使用 `ILIKE` / `lower()` 比较 |
| 保留字表名 | `user` | `admin_user` |

迁移脚本可重复执行（`IF NOT EXISTS` / `CREATE OR REPLACE` / 先 `DROP TRIGGER IF EXISTS`），
一键启动脚本每次启动都会幂等执行一次。

### 迁移是怎么被执行的（两层）

| 层 | 触发时机 | 位置 | 行为 |
| --- | --- | --- | --- |
| ① 镜像初始化 | **仅在数据卷为空时**（`PGDATA` 下没有 `PG_VERSION`） | `docker-compose.yaml` 把 DDL 挂到 `/docker-entrypoint-initdb.d/001_init.sql` | 官方镜像 entrypoint 用 `psql -f` 执行一次 |
| ② 一键脚本 | **每次 `start` / `restart` / `migrate`** | `scripts/run-pyadmin.sh` 的 `migrate()`（第 28-38 行） | 等 postgres 健康后用 `psql -q -v ON_ERROR_STOP=1 -f` 重放同一份 DDL，成功后才启动 pyadmin |

因此 `./scripts/run-pyadmin.sh start` 对**新卷和旧卷都会**执行迁移；而直接 `docker compose up`
（不走脚本）只有第 ① 层，已有数据卷不会补表，需要显式执行 `./scripts/run-pyadmin.sh migrate`。

幂等边界：只覆盖“新增式”变更（`CREATE ... IF NOT EXISTS`、`CREATE OR REPLACE FUNCTION`、
先 `DROP TRIGGER IF EXISTS` 再建触发器）。**改名、删列、改类型这类破坏性变更不会自动生效**，
需要在 `resources/migrations` 追加升级脚本并手动执行，本地可直接用 `reset` 重建空库。

自查命令（表 6 张、触发器 6 个、索引 26 个）：

```bash
docker compose -f pyadmin/docker-compose.yaml exec -T postgres \
  psql -U hercules -d hercules_admin -c '\dt' \
  -c "select count(*) from pg_trigger where not tgisinternal"
```

## 配置

配置位于 `pyadmin/config`，按 `APP_ENV` 选择文件，未设置时默认 `dev`（与 Go 版本一致，不做多文件合并）。

| 文件 | APP_ENV | 用途 |
| --- | --- | --- |
| `config.yaml` | `default` | 默认基础配置 |
| `config.dev.yaml` | `dev` | 本地开发（`127.0.0.1:5432`） |
| `config.test.yaml` | `test` | 自动化测试（`hercules_admin_test`） |
| `config.prod.yaml` | `prod` | Compose / 生产（连 `postgres` 服务） |

环境变量覆盖（优先级高于 YAML）：

| 变量 | 作用 | 默认 |
| --- | --- | --- |
| `APP_ENV` | 选择环境配置文件 | `dev` |
| `DATABASE_DSN` | 数据库连接串，支持 `postgres://`、`postgresql://` 自动转为 asyncpg | 取配置文件 |
| `JWT_SECRET` | JWT 签名密钥（生产必须替换） | 取配置文件 |
| `JWT_EXPIRE_HOURS` | 令牌有效小时数 | `24`（prod 为 `12`） |
| `SERVER_HOST` / `SERVER_PORT` | 监听地址与端口 | `0.0.0.0` / `8000` |
| `PYADMIN_CONFIG_DIR` | 配置文件目录（容器内为 `/app/config`） | `pyadmin/config` |

## 快速开始（Docker 一键启动）

```bash
cp pyadmin/.env.example pyadmin/.env    # 可选，可先修改密码与 JWT_SECRET
./scripts/run-pyadmin.sh start
```

启动后：

- 管理台：`http://127.0.0.1:5173`
- 后端 API：`http://127.0.0.1:8000`
- 接口文档：`http://127.0.0.1:8000/docs`
- PostgreSQL：`127.0.0.1:5432`（库 `hercules_admin`）
- 默认超级管理员：`admin` / `admin`（UID `1000000000`，Org ID `10000000`）

脚本动作：

```bash
./scripts/run-pyadmin.sh start            # 起库、幂等迁移、构建并启动全部服务
./scripts/run-pyadmin.sh stop             # 停止全部容器
./scripts/run-pyadmin.sh restart
./scripts/run-pyadmin.sh build
./scripts/run-pyadmin.sh migrate          # 仅执行数据库迁移（幂等）
./scripts/run-pyadmin.sh logs pyadmin     # 跟踪指定服务日志
./scripts/run-pyadmin.sh status
./scripts/run-pyadmin.sh test             # 运行 pytest
./scripts/run-pyadmin.sh reset            # 删除数据卷重建空库（需输入 yes 确认）
```

`docker-compose.yaml` 定义三个服务：`postgres`（16-alpine，带健康检查与初始化 DDL）、`pyadmin`（FastAPI）、
`admin-web`（复用 `../admin-web` 镜像，挂载 `deploy/nginx.conf` 把 `/api` 反代到 `pyadmin:8000`），
因此前端代码零改动。

## 本地开发

```bash
# 1. 数据库
docker compose -f pyadmin/docker-compose.yaml up -d --wait postgres

# 2. 后端
cd pyadmin
python3 -m venv .venv && . .venv/bin/activate
pip install -r requirements-dev.txt
python -m app                     # 默认 APP_ENV=dev，监听 8000

# 3. 前端
cd ../admin-web && npm install && npm run dev   # Vite 将 /api 代理到 127.0.0.1:8000
```

## 测试

```bash
cd pyadmin
python -m pytest                  # 单元测试 + 全链路接口测试（SQLite，秒级）
ruff check app tests --line-length 120
```

- `tests/test_units.py`：ID 生成、bcrypt、JWT（过期/错误密钥/缺失声明）、状态枚举、分页、响应封装、
  时间格式、权限编码正则、校验提示、配置加载与环境变量覆盖。
- `tests/test_auth_profile_api.py`、`test_organization_api.py`、`test_user_api.py`、
  `test_role_permission_api.py`：覆盖登录注册、个人中心、组织树与环路校验、用户/角色/权限 CRUD、
  乐观锁冲突、唯一性冲突、组织隔离、直接权限与角色互斥清理、用户与组织禁用即时失效等场景。
- `tests/test_postgres_integration.py`：真实 PostgreSQL 校验（建表、索引、注释、`updated_at` 触发器、
  完整业务链路）。默认跳过，指定专用测试库后执行：

```bash
PYADMIN_TEST_PG_DSN=postgresql+asyncpg://hercules:hercules@127.0.0.1:5432/hercules_admin_test \
    python -m pytest tests/test_postgres_integration.py
```

该用例会先删除 6 张表再重放迁移脚本，因此**库名必须包含 `test`**，请勿指向业务库。

## 注意事项

- `docker-entrypoint-initdb.d` 只在数据卷首次创建时执行；脚本每次 `start` 都会幂等再执行一次 DDL，
  因此已有旧卷也能补齐缺失的表。改动 DDL 后如要完全重建，用 `./scripts/run-pyadmin.sh reset`。
- 生产环境务必替换 `JWT_SECRET`，并建议在反向代理层启用 HTTPS、收敛 CORS 域名（`server.corsOrigins`）。
- 每个请求都会重新校验用户与组织状态：禁用用户或禁用所属组织后，已签发的 token 立即失效（401）。
- 用户直接权限通过 `role_id = -uid` 的隐藏个人角色承载；重新分配角色会清空直接权限，
  跨组织移动用户会清空其角色与直接权限，避免权限跨组织泄漏。
- 组织管理（增删改）业务上仅超级管理员可执行；组织删除前会校验子组织、用户、角色、权限是否为空。
- 权限只在所属组织内生效，不继承父组织或子组织；权限编码需为 `report.read` 形式的小写点分格式。
- 接口错误一律返回统一响应体，不泄露数据库与堆栈细节；日志在服务端输出。
