"""命令行入口：python -m app 启动 HTTP 服务。"""

from __future__ import annotations

import uvicorn

from app.config import load_settings


def main() -> None:
    """按当前环境配置启动 uvicorn。"""

    settings = load_settings()
    uvicorn.run(
        "app.main:app",
        host=settings.server.host,
        port=settings.server.port,
        log_level="info",
        proxy_headers=True,
    )


if __name__ == "__main__":
    main()
