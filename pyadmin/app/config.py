"""配置加载：按 APP_ENV 选择 config.<env>.yaml，并支持环境变量覆盖。

与 Go 版本保持一致：未显式指定 APP_ENV 时默认使用 dev 环境；不同环境使用独立配置文件，
不做多层合并，避免环境之间隐式继承。
"""

from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Tuple

import yaml

DEFAULT_ENVIRONMENT = "dev"

SUPPORTED_ENVIRONMENTS = ("default", "dev", "test", "prod")

_DEFAULT_CONFIG_DIR = Path(__file__).resolve().parent.parent / "config"


@dataclass(frozen=True)
class ServerSettings:
    """HTTP 服务配置。"""

    host: str
    port: int
    cors_origins: Tuple[str, ...]


@dataclass(frozen=True)
class DatabaseSettings:
    """数据库连接配置。"""

    dsn: str
    echo: bool
    pool_size: int
    max_overflow: int
    pool_recycle_seconds: int


@dataclass(frozen=True)
class SecuritySettings:
    """安全配置。"""

    jwt_secret: str
    jwt_expire_hours: int

    @property
    def expires_in_seconds(self) -> int:
        """令牌有效秒数，用于登录响应。"""

        return self.jwt_expire_hours * 3600


@dataclass(frozen=True)
class Settings:
    """运行期配置聚合。"""

    environment: str
    server: ServerSettings
    database: DatabaseSettings
    security: SecuritySettings
    config_file: Path


def _config_dir() -> Path:
    override = os.getenv("PYADMIN_CONFIG_DIR", "").strip()
    return Path(override) if override else _DEFAULT_CONFIG_DIR


def resolve_environment(environment: str | None = None) -> str:
    """解析并校验运行环境。"""

    value = (environment or os.getenv("APP_ENV") or "").strip() or DEFAULT_ENVIRONMENT
    if value not in SUPPORTED_ENVIRONMENTS:
        raise ValueError("不支持的 APP_ENV: {}".format(value))
    return value


def _read_yaml(path: Path) -> Dict[str, Any]:
    if not path.is_file():
        raise FileNotFoundError("配置文件不存在: {}".format(path))
    with path.open("r", encoding="utf-8") as handle:
        payload = yaml.safe_load(handle) or {}
    if not isinstance(payload, dict):
        raise ValueError("配置文件格式无效: {}".format(path))
    return payload


def _section(payload: Dict[str, Any], name: str) -> Dict[str, Any]:
    value = payload.get(name) or {}
    if not isinstance(value, dict):
        raise ValueError("配置项 {} 必须是对象".format(name))
    return value


def _as_bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    return str(value).strip().lower() in {"1", "true", "yes", "on"}


def _as_int(value: Any, fallback: int) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return fallback


def _normalize_dsn(raw: str) -> str:
    """把常见的 postgres DSN 规范化为 SQLAlchemy asyncpg 形式。"""

    value = raw.strip()
    for prefix in ("postgres://", "postgresql://"):
        if value.startswith(prefix):
            return "postgresql+asyncpg://" + value[len(prefix) :]
    return value


def load_settings(environment: str | None = None) -> Settings:
    """加载指定环境的配置。"""

    resolved = resolve_environment(environment)
    file_name = "config.yaml" if resolved == "default" else "config.{}.yaml".format(resolved)
    config_file = _config_dir() / file_name
    payload = _read_yaml(config_file)

    server = _section(payload, "server")
    database = _section(payload, "database")
    security = _section(payload, "security")

    origins = server.get("corsOrigins") or ["*"]
    if isinstance(origins, str):
        origins = [origins]

    dsn = _normalize_dsn(os.getenv("DATABASE_DSN") or str(database.get("dsn") or ""))
    if not dsn:
        raise ValueError("数据库 DSN 不能为空")

    secret = os.getenv("JWT_SECRET") or str(security.get("jwtSecret") or "")
    if not secret:
        raise ValueError("JWT 密钥不能为空")

    host = os.getenv("SERVER_HOST") or str(server.get("host") or "0.0.0.0")
    port = _as_int(os.getenv("SERVER_PORT") or server.get("port"), 8000)
    expire_hours = _as_int(
        os.getenv("JWT_EXPIRE_HOURS") or security.get("jwtExpireHours"), 24
    )

    return Settings(
        environment=resolved,
        server=ServerSettings(
            host=host,
            port=port,
            cors_origins=tuple(str(item) for item in origins),
        ),
        database=DatabaseSettings(
            dsn=dsn,
            echo=_as_bool(database.get("echo", False)),
            pool_size=_as_int(database.get("poolSize"), 5),
            max_overflow=_as_int(database.get("maxOverflow"), 10),
            pool_recycle_seconds=_as_int(database.get("poolRecycleSeconds"), 1800),
        ),
        security=SecuritySettings(
            jwt_secret=secret,
            jwt_expire_hours=max(expire_hours, 1),
        ),
        config_file=config_file,
    )
