## Hướng dẫn migration

### 1) Cài đặt lần đầu (chỉ chạy 1 lần)

```bash
chmod +x install.sh migration.sh
./install.sh
```

Script `install.sh` sẽ tạo virtualenv tại `.venv/` và cài dependencies trong `requirements.txt`.

### 2) Sử dụng migration.sh

Script `migration.sh` chỉ thực thi Alembic. Nếu `.venv/` tồn tại nó sẽ tự kích hoạt, không tự cài đặt thêm gì.

-  định (không tham số): đảm bảo có revision đầu tiên (nếu chưa có) và `upgrade head`.

```bash
./migration.sh
```

- Truyền tham số Alembic trực tiếp (pass-through):

```bash
# Tạo revision tự động với message
./migration.sh revision --autogenerate -m "add users table"

# Nâng cấp/rollback
./migration.sh upgrade head
./migration.sh downgrade -1
```

### 3) Cấu hình DATABASE_URL

`migration.sh` dùng biến môi trường `DATABASE_URL`. Nếu không đặt, mặc định là SQLite tại `./test.db`.

Ví dụ:

```bash
# SQLite (mặc định)
./migration.sh

# Postgres
export DATABASE_URL="postgresql+psycopg2://user:pass@localhost:5432/mydb"
./migration.sh upgrade head

# MySQL
export DATABASE_URL="mysql+pymysql://user:pass@localhost:3306/mydb"
./migration.sh upgrade head
```

### 4) Cấu trúc liên quan

- `migration/alembic.ini`: cấu hình Alembic
- `migration/alembic/env.py`: load `migration/schema.py` để autogenerate
- `migration/alembic/versions/`: nơi chứa các file revision

### 5) Lưu ý

- Lần đầu clone: chạy `./install.sh` trước khi dùng `./migration.sh`.
- Khi thay đổi model trong `migration/schema.py`, tạo revision mới bằng `./migration.sh revision --autogenerate -m "message"` rồi `./migration.sh upgrade head`.
