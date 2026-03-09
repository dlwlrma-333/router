# API Reference（开发必读）

本文件是 Router 项目的 **API 开发规范与总览**。  
凡涉及 API 新增/修改/迁移，请先阅读并遵循本文件。

## 1) 公司 Interface 规范（必须遵守）
每个服务的接口 **只能属于一种类型**：
- **public（公开接口）**：面向第三方、前端/移动端、外部开发者  
- **admin（运营接口）**：内部运营/客服/风控或企业租户管理员  
- **internal（内部接口）**：服务间调用、批处理/定时任务、内部系统  

选择规则：
- 只要包含公开接口能力 → **public**
- 覆盖运营+内部 → **admin**
- 仅服务/Job 调用 → **internal**

示例：
- `GET /api/v1/public/products/{id}`
- `POST /api/v1/admin/users/{id}:block`
- `POST /api/v1/internal/search:index`

规范原文（链接仅供参考）：
```
https://wiki.yeying.pub/link/266#bkmrk-%E5%85%BC%E5%AE%B9%E6%80%A7%EF%BC%8C%E5%B0%BD%E9%87%8F%E4%B8%8D%E5%8A%A8%E8%B7%AF%E7%BB%8F%E7%9A%84%E7%89%88%E6%9C%AC%E5%8F%B7%2Fapi%2Fv
```

## 2) 当前 Router 的对外 API 结构
**规范路径：**
- `/api/v1/public/*`
- `/api/v1/admin/*`
- `/api/v1/internal/*`（当前预留）

**兼容路径（历史保留）：**
- `/api/*`（旧管理/用户接口）
- `/v1/*` 与 `/dashboard/*`（OpenAI 兼容）

**迁移原则：**
- 新增/改动 **只做在 /api/v1/**  
- 旧路径仅做镜像兼容，不再新增功能

## 3) 鉴权与权限
本项目对外推荐 **JWT**：
```
Authorization: Bearer <JWT>
```

角色/权限：
- public 用户侧：`UserAuth`（普通用户 JWT）
- admin 管理侧：`AdminAuth`（管理员 JWT）
- root 配置：`RootAuth`（Root JWT）

OpenAI 兼容调用（`/api/v1/public/*`）使用 `TokenAuth + Distribute`，与旧 `/v1/*` 行为一致。

## 4) 兼容性与版本策略
接口变更需要判断是否 **向下兼容**：
- 能兼容：**不改** `/api/v1/*` 版本号  
- 不兼容：新路径升级到 `/api/v2/*`

## 5) 接口定义与存放位置（规范要求）
接口应在 **interface 库** 中以 proto 形式定义（支持 grpc/http）。  
新接口应放置于：
- `yeying/api/common`（公共通用接口）
- `yeying/api/<app name>`（服务私有接口）

> 本仓库目前以 Gin 路由为主，但新增 API 仍应对齐该规范与分层约束。

## 6) 开发落地指引（Router 实际代码）
涉及路由与中间件的主要入口：
- 路由注册：`internal/transport/http/router/api.go`
- 鉴权：`internal/transport/http/middleware/auth.go`
- OpenAI 兼容：`internal/transport/http/router/relay.go`

新增/改动 API 时，请同步更新文档：
- `docs/接口文档.md`（统一 API 文档）
- `docs/swagger/openapi.json`（Swagger/OpenAPI 产物，执行 `scripts/gen-openapi.sh` 生成）

Swagger 注释要求：
- 为 `/api/v1/public` 与 `/api/v1/admin` handler 补充 swag 注释（不要写旧 `/api` 或 `/v1`）
- 生成 OpenAPI：`scripts/gen-openapi.sh`

## 7) 迁移与禁用开关（供运维/发布）
本仓库支持环境变量：
- `DISABLE_OPENAI_COMPAT=true` 可禁用 `/v1/*` 与 `/dashboard/*`  
  仅保留 `/api/v1/public/*` 作为对外入口。

---

如需新增 API，请先确定分层类型并写清楚 **public/admin/internal** 的归属与理由，再动代码。
