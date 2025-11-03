from alembic import command
from alembic.config import Config
from alembic.environment import EnvironmentContext
from alembic.script import ScriptDirectory

def run_migrations_online():
    config = Config("alembic.ini")
    config.set_main_option("sqlalchemy.url", "sqlite:///test.db")
    with EnvironmentContext(config, target_metadata=target_metadata) as context:
        script = ScriptDirectory.from_config(config)
        runner = MigrationContext.configure(connection, target_metadata=target_metadata)
        runner.run(script, revision="head")