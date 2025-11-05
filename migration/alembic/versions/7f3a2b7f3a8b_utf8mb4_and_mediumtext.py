"""utf8mb4 charset and MEDIUMTEXT for chunks.content

Revision ID: 7f3a2b7f3a8b
Revises: f9581994c712
Create Date: 2025-11-05 09:55:00.000000

"""
from alembic import op


# revision identifiers, used by Alembic.
revision = "7f3a2b7f3a8b"
down_revision = "f9581994c712"
branch_labels = None
depends_on = None


def upgrade() -> None:
    bind = op.get_bind()
    dialect = bind.dialect.name if bind is not None else ""

    if dialect == "mysql":
        # Ensure all tables use utf8mb4 (handles existing rows too)
        op.execute("ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;")
        op.execute("ALTER TABLE documents CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;")
        op.execute("ALTER TABLE messages CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;")
        op.execute("ALTER TABLE chunks CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;")

        # Adjust columns on chunks
        # - content -> MEDIUMTEXT for larger payloads
        # - content_preview stays VARCHAR(512) but ensure utf8mb4 to avoid Error 1366
        # - milvus_collection ensure utf8mb4
        # - content_hash is hex string -> ascii for compactness (optional)
        op.execute("ALTER TABLE chunks MODIFY content MEDIUMTEXT CHARACTER SET utf8mb4;")
        op.execute("ALTER TABLE chunks MODIFY content_preview VARCHAR(512) CHARACTER SET utf8mb4;")
        op.execute("ALTER TABLE chunks MODIFY milvus_collection VARCHAR(128) CHARACTER SET utf8mb4;")
        op.execute("ALTER TABLE chunks MODIFY content_hash CHAR(64) CHARACTER SET ascii;")
    else:
        # No-op for non-MySQL dialects
        pass


def downgrade() -> None:
    bind = op.get_bind()
    dialect = bind.dialect.name if bind is not None else ""

    if dialect == "mysql":
        # Revert column changes on chunks
        op.execute("ALTER TABLE chunks MODIFY content TEXT CHARACTER SET utf8mb4;")
        op.execute("ALTER TABLE chunks MODIFY content_preview VARCHAR(512) CHARACTER SET utf8mb4;")
        op.execute("ALTER TABLE chunks MODIFY milvus_collection VARCHAR(128) CHARACTER SET utf8mb4;")
        op.execute("ALTER TABLE chunks MODIFY content_hash VARCHAR(64) CHARACTER SET utf8mb4;")
        # Keep utf8mb4 conversion on tables; reverting collations may cause data loss, so we skip.
    else:
        # No-op for non-MySQL dialects
        pass


