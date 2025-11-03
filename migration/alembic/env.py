import os
import sys
from typing import Any, Dict
from logging.config import fileConfig

from sqlalchemy import engine_from_config, pool
from alembic import context
try:
    import yaml  # type: ignore
except Exception:
    yaml = None  # will fall back if missing

# this is the Alembic Config object, which provides
# access to the values within the .ini file in use.
config = context.config

# Interpret the config file for Python logging.
# This line sets up loggers basically.
if config.config_file_name is not None:
    fileConfig(config.config_file_name)

# Ensure project root is on sys.path so we can import migration.schema
PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
if PROJECT_ROOT not in sys.path:
    sys.path.insert(0, PROJECT_ROOT)

# Import metadata from our models
from migration.schema import Base  # noqa: E402

target_metadata = Base.metadata


def _load_yaml_config(path: str) -> Dict[str, Any]:
    if yaml is None:
        return {}
    if not os.path.exists(path):
        return {}
    with open(path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f) or {}


def _db_url_from_yaml(cfg: Dict[str, Any]) -> str | None:
    # Support both keys: "database" (Go) and "db" (current yaml)
    db_cfg = cfg.get("database") or cfg.get("db") or {}

    # If explicit url is provided, use it directly
    url = db_cfg.get("url")
    if isinstance(url, str) and url.strip():
        return url.strip()

    host = db_cfg.get("host")
    port = db_cfg.get("port")
    user = db_cfg.get("user")
    password = db_cfg.get("password")
    name = db_cfg.get("name")
    driver = db_cfg.get("driver")  # e.g., postgresql+psycopg2, mysql+pymysql, sqlite

    # If nothing useful, return None to allow fallback
    if not any([host, user, password, name, driver]):
        return None

    # Default to MySQL (PyMySQL) if not specified
    driver = (driver or "mysql+pymysql").strip()

    if driver.startswith("sqlite"):
        # File-based sqlite when name is provided, otherwise in-memory
        if isinstance(name, str) and name:
            # relative to repo root
            db_path = os.path.join(PROJECT_ROOT, name)
            return f"sqlite:///{db_path}"
        return "sqlite://"

    # For networked DBs
    auth = ""
    if isinstance(user, str) and user:
        if isinstance(password, str) and password:
            auth = f"{user}:{password}@"
        else:
            auth = f"{user}@"
    host_part = str(host) if host else "localhost"
    port_part = f":{port}" if port else ""
    name_part = f"/{name}" if name else ""
    return f"{driver}://{auth}{host_part}{port_part}{name_part}"


def get_database_url() -> str:
    # 1) Prefer config.yaml if present
    yaml_path = os.path.join(PROJECT_ROOT, "config.yaml")
    cfg = _load_yaml_config(yaml_path)
    url_from_yaml = _db_url_from_yaml(cfg)
    if url_from_yaml:
        return url_from_yaml

    # 2) Then env var
    env_url = os.getenv("DATABASE_URL")
    if env_url:
        return env_url

    # 3) Fallback to alembic.ini setting
    return config.get_main_option("sqlalchemy.url")


def run_migrations_offline() -> None:
    url = get_database_url()
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
        compare_type=True,
        compare_server_default=True,
    )

    with context.begin_transaction():
        context.run_migrations()


def run_migrations_online() -> None:
    configuration = config.get_section(config.config_ini_section) or {}
    configuration["sqlalchemy.url"] = get_database_url()

    connectable = engine_from_config(
        configuration,
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    with connectable.connect() as connection:
        context.configure(
            connection=connection,
            target_metadata=target_metadata,
            compare_type=True,
            compare_server_default=True,
        )

        with context.begin_transaction():
            context.run_migrations()


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()


