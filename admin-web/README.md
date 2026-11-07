# Hercules Admin Web

React + TypeScript + Vite 管理台，采用明亮、低饱和视觉体系，包含响应式桌面与移动布局。

## 页面

- 登录与注册（按租户名称或租户编码搜索，登录时提交 `tenant_code`）
- 运营概览
- 租户树 CRUD
- 用户 CRUD、租户成员与角色/直接权限分配
- 角色 CRUD 与角色权限分配
- 权限 CRUD 与 API 绑定
- 个人资料、角色和权限查看

新建租户、用户和角色时无需输入业务 ID。管理员创建用户时可留空密码，由后端生成一次性初始密码；结果弹窗支持查看。权限节点类型支持 `MENU`、`BUTTON`、`API`、`DATA`，编码前缀分别为 `menu.`、`button.`、`api.`、`data.`。

## 开发

```bash
npm install
npm run dev
```

开发服务为 `http://127.0.0.1:5173`，`/api` 自动代理到 `http://127.0.0.1:8000`。

## 生产构建

```bash
npm run build
```

产物位于 `dist/`。生产部署时应由 Web 服务器托管静态文件，并将 `/api` 反向代理到后端服务。

项目根目录的 `docker-compose.yaml` 会构建本目录的多阶段 Dockerfile，最终由 Nginx 提供静态文件并将 `/api` 代理到 `admin:8000`。可使用 `./scripts/run.sh start` 与其他服务统一启动。
