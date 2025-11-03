# app/models.py
from sqlalchemy.orm import declarative_base, relationship
from sqlalchemy import (
    Column, BigInteger, Integer, String, Text, Enum,
    ForeignKey, TIMESTAMP, func
)

Base = declarative_base()

class User(Base):
    __tablename__ = "users"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    email = Column(String(255), unique=True, nullable=False)
    username = Column(String(100), unique=True)
    password_hash = Column(String(255))
    created_at = Column(TIMESTAMP, server_default=func.current_timestamp())

class Document(Base):
    __tablename__ = "documents"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    title = Column(String(255))
    original_filename = Column(String(255))
    file_path = Column(String(500))
    language = Column(String(20), server_default="en")
    page_count = Column(Integer)
    sha256 = Column(String(64), unique=True)
    uploaded_at = Column(TIMESTAMP, server_default=func.current_timestamp())

class Chunk(Base):
    __tablename__ = "chunks"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    document_id = Column(BigInteger, ForeignKey("documents.id", ondelete="CASCADE"), nullable=False)
    chunk_index = Column(Integer, nullable=False)
    page_index = Column(Integer)
    content = Column(Text, nullable=False)            # MEDIUMTEXT sẽ chỉnh trong migration
    content_preview = Column(String(512))
    token_count = Column(Integer)
    milvus_collection = Column(String(128), nullable=False)
    milvus_id = Column(BigInteger, nullable=False)
    content_hash = Column(String(64), nullable=False)
    created_at = Column(TIMESTAMP, server_default=func.current_timestamp())

class Message(Base):
    __tablename__ = "messages"
    id = Column(BigInteger, primary_key=True, autoincrement=True)
    user_id = Column(BigInteger, ForeignKey("users.id", ondelete="CASCADE"), nullable=False)
    role = Column(Enum("user", "assistant", name="role_enum"), nullable=False)
    content = Column(Text, nullable=False)
    document_id = Column(BigInteger, ForeignKey("documents.id", ondelete="SET NULL"))
    created_at = Column(TIMESTAMP, server_default=func.current_timestamp())
