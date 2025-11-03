## Hướng dẫn migration

### 1) Cài đặt lần đầu (chỉ chạy 1 lần)

```bash
chmod +x install.sh migration.sh
./install.sh
```

Script `install.sh` sẽ tạo virtualenv tại `.venv/` và cài dependencies trong `requirements.txt`.

### 2) Sử dụng migration.sh

Script `migration.sh` chỉ thực thi Alembic. Nếu `.venv/` tồn tại nó sẽ tự kích hoạt, không tự cài đặt thêm gì.

- Mặc định (không tham số): đảm bảo có revision đầu tiên (nếu chưa có) và `upgrade head`.

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

### 3) Nguồn cấu hình DB (luôn lấy từ config.yaml)

`env.py` sẽ luôn đọc file `config.yaml` ở thư mục gốc để dựng `DATABASE_URL` theo thứ tự ưu tiên:

1. `config.yaml` (khóa `database` hoặc `db`):
   - Nếu có `url`: dùng trực tiếp.
   - Nếu không, sẽ ghép từ `driver` (mặc định `mysql+pymysql`), `host`, `port`, `user`, `password`, `name`.
   - Nếu `driver` là `sqlite`, tên DB sẽ được coi là đường dẫn tệp trong repo.
2. Biến môi trường `DATABASE_URL` (fallback nếu thiếu thông tin trong `config.yaml`).
3. `alembic.ini` (cuối cùng).

Ví dụ `config.yaml`:

```yaml
db:
  # Tuỳ chọn 1: truyền URL trực tiếp
  # url: mysql+pymysql://user:pass@localhost:3306/ai-learn-english

  # Tuỳ chọn 2: truyền từng trường (driver mặc định mysql+pymysql)
  host: localhost
  port: 2500
  user: user
  password: password
  name: ai-learn-english
  # driver: mysql+pymysql
```

### 4) Cấu trúc liên quan

- `migration/alembic.ini`: cấu hình Alembic
- `migration/alembic/env.py`: load `migration/schema.py` để autogenerate
- `migration/alembic/versions/`: nơi chứa các file revision

### 5) Lưu ý

- Lần đầu clone: chạy `./install.sh` trước khi dùng `./migration.sh`.
- Khi thay đổi model trong `migration/schema.py`, tạo revision mới bằng `./migration.sh revision --autogenerate -m "message"` rồi `./migration.sh upgrade head`.
