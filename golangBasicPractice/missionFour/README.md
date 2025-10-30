# Blog API - Mission Four

基于 Go + Gin + GORM + MySQL 的博客 API 系统。

## 功能特性

- 用户注册/登录（JWT 认证）
- 文章 CRUD 操作
- 评论系统
- 权限控制（只能操作自己的文章）

## 环境要求

- Go 1.19+
- MySQL 8.0+

## 安装步骤

### 1. 安装依赖

**Go:**
```bash
# macOS
brew install go

# Ubuntu
sudo apt install golang-go
```

**MySQL:**
```bash
# macOS
brew install mysql
brew services start mysql

# Ubuntu
sudo apt install mysql-server
sudo systemctl start mysql
```

### 2. 配置数据库

登录 MySQL 创建数据库：
```sql
mysql -u root -p
CREATE DATABASE four CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
EXIT;
```

### 3. 配置环境变量

```bash
export DB_HOST=127.0.0.1
export DB_PORT=3306
export DB_NAME=four
export DB_USER=root
export DB_PASSWORD=your_mysql_password
```

### 4. JWT 配置

创建 `configs/config.yaml`：
```yaml
jwt:
  secret: "your-jwt-secret-key"
  expiry: "24h"
```

## 启动项目

```bash
cd missionFour
go mod tidy
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API 接口

### 认证
```bash
# 注册
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com","password":"123456"}'

# 登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"123456"}'
```

### 文章管理（需要 Authorization: Bearer <token>）
```bash
# 创建文章
curl -X POST http://localhost:8080/api/post/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"My Post","content":"Content here","user_id":1}'

# 获取文章列表
curl -X GET http://localhost:8080/api/post/list \
  -H "Authorization: Bearer <token>"

# 更新文章
curl -X PUT http://localhost:8080/api/post/update \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"id":1,"title":"Updated Title","content":"Updated content","user_id":1}'

# 删除文章
curl -X DELETE http://localhost:8080/api/post/delete?postID=1 \
  -H "Authorization: Bearer <token>"
```

### 评论（需要 Authorization: Bearer <token>）
```bash
# 添加评论
curl -X POST http://localhost:8080/api/comment/insert \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"post_id":1,"content":"Great post!"}'

# 获取评论
curl -X GET "http://localhost:8080/api/comment/get?postID=1" \
  -H "Authorization: Bearer <token>"
```

## Postman 测试

导入 `Blog_API_Postman_Collection.json` 文件到 Postman 进行测试。

## 常见问题

1. **数据库连接失败**：检查 MySQL 服务是否运行，配置是否正确
2. **JWT 错误**：重新登录获取新 token
3. **权限错误**：确保只能操作自己的文章

## 项目结构

```
missionFour/
├── main.go                 # 程序入口
├── source/DataBase.go     # 数据库操作
├── router/                # 路由处理
├── internal/              # 内部包
└── configs/               # 配置文件
```