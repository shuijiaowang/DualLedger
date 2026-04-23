# Go-Vue-Demo

最小 **Gin + GORM(MySQL) + Vue3(Vite)** 模版：用户注册/登录、JWT、占位 `example` 表与鉴权示例接口。

## 目录

- `service/`：Go 后端（模块名 `SService`，可按项目改名并 `go mod tidy`）。
- `web/`：Vue 前端。

## 后端

1. 创建数据库（默认名见 `service/config/config.yaml` 的 `mysql.dbname`，如 `go_vue_demo`）。
2. 修改 `service/config/config.yaml`：`mysql`、`jwt.secret`、`server.port`。
3. 在 `service` 目录执行：

```bash
go mod tidy
go run .
```

HTTP 端口来自 **`server.port`**，不要用硬编码。

## 前端

```bash
cd web
npm install
npm run dev
```

开发时 Vite 将 `VITE_BASE_API`（默认 `/api`）代理到 `VITE_BASE_PATH` + `VITE_SERVER_PORT`，需与后端 `server.port` 一致（见 `web/.env.development`）。

## 数据表

GORM `AutoMigrate` 仅包含 **`user`**、**`example`**（均嵌入 `gorm.Model`）。模型在 `service/model/`。

## API 摘要

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/user/register` | 注册 |
| POST | `/api/user/login` | 登录，返回 JWT |
| POST | `/api/example/test` | 需 `Authorization: Bearer <token>` |
