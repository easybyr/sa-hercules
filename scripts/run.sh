#!/usr/bin/env sh

set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
compose_file="$project_dir/docker-compose.yaml"
action=${1:-start}

case "$action" in
  start)
    docker compose -f "$compose_file" up -d --build --wait
    echo "Hercules Admin 已启动：http://127.0.0.1:${ADMIN_WEB_PORT:-5173}"
    ;;
  stop)
    docker compose -f "$compose_file" down
    ;;
  restart)
    docker compose -f "$compose_file" down
    docker compose -f "$compose_file" up -d --build --wait
    ;;
  build)
    docker compose -f "$compose_file" build
    ;;
  logs)
    if [ "$#" -ge 2 ]; then
      docker compose -f "$compose_file" logs -f "$2"
    else
      docker compose -f "$compose_file" logs -f
    fi
    ;;
  status)
    docker compose -f "$compose_file" ps
    ;;
  *)
    echo "用法: $0 {start|stop|restart|build|logs [service]|status}" >&2
    exit 2
    ;;
esac
