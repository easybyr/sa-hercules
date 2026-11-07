#!/usr/bin/env sh
# PyAdmin 一键启动脚本：PostgreSQL + pyadmin(FastAPI) + admin-web(React/Nginx)
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
compose_file="$project_dir/pyadmin/docker-compose.yaml"
env_file="$project_dir/pyadmin/.env"
action=${1:-start}

ensure_env() {
  if [ ! -f "$env_file" ]; then
    cp "$project_dir/pyadmin/.env.example" "$env_file"
    echo "已生成 $env_file，可按需修改后重试"
  fi
}

compose() {
  docker compose --env-file "$env_file" -f "$compose_file" "$@"
}

env_value() {
  key=$1
  fallback=$2
  value=$(sed -n "s/^${key}=//p" "$env_file" 2>/dev/null | tail -n 1)
  if [ -n "${value:-}" ]; then printf '%s' "$value"; else printf '%s' "$fallback"; fi
}

migrate() {
  ensure_env
  db_user=$(env_value POSTGRES_USER hercules)
  db_name=$(env_value POSTGRES_DB hercules_admin)
  echo "应用数据库迁移（脚本幂等，可重复执行）..."
  # 已存在的对象会输出 NOTICE，这里降到 warning 级避免误读为报错；ERROR 仍会打印并使脚本失败。
  compose exec -T -e PGOPTIONS="-c client_min_messages=warning" postgres \
    psql -q -v ON_ERROR_STOP=1 -U "$db_user" -d "$db_name" \
    -f /docker-entrypoint-initdb.d/001_init.sql >/dev/null
  echo "迁移完成"
}

start() {
  ensure_env
  echo "启动 PostgreSQL 并等待健康检查..."
  compose up -d --wait postgres
  migrate
  echo "构建并启动 pyadmin 与 admin-web..."
  compose up -d --build --wait
  web_port=$(env_value ADMIN_WEB_PORT 5173)
  api_port=$(env_value PYADMIN_PORT 8000)
  printf '\nPyAdmin 已启动：\n'
  printf '  管理台：http://127.0.0.1:%s\n' "$web_port"
  printf '  后端 API：http://127.0.0.1:%s\n' "$api_port"
  printf '  接口文档：http://127.0.0.1:%s/docs\n' "$api_port"
  printf '  PostgreSQL：127.0.0.1:%s\n' "$(env_value POSTGRES_PORT 5432)"
  printf '  默认超级管理员：admin / admin\n'
}

run_tests() {
  venv_python="$project_dir/pyadmin/.venv/bin/python"
  if [ -x "$venv_python" ]; then
    (cd "$project_dir/pyadmin" && "$venv_python" -m pytest)
  elif command -v python3 >/dev/null 2>&1; then
    echo "未检测到 pyadmin/.venv，使用系统 python3（需先安装 requirements-dev.txt）"
    (cd "$project_dir/pyadmin" && python3 -m pytest)
  else
    echo "未找到可用的 Python，请先创建虚拟环境并安装 pyadmin/requirements-dev.txt" >&2
    exit 2
  fi
}

case "$action" in
  start)
    start
    ;;
  stop)
    ensure_env
    compose down
    ;;
  restart)
    ensure_env
    compose down
    start
    ;;
  build)
    ensure_env
    compose build
    ;;
  migrate)
    migrate
    ;;
  logs)
    ensure_env
    if [ "$#" -ge 2 ]; then compose logs -f "$2"; else compose logs -f; fi
    ;;
  status)
    ensure_env
    compose ps
    ;;
  test)
    run_tests
    ;;
  reset)
    ensure_env
    printf '该操作会删除 PostgreSQL 数据卷并重建空库，输入 yes 继续：'
    read -r answer
    if [ "$answer" != "yes" ]; then
      echo "已取消"
      exit 1
    fi
    compose down -v
    start
    ;;
  *)
    echo "用法: $0 {start|stop|restart|build|migrate|logs [service]|status|test|reset}" >&2
    exit 2
    ;;
esac
