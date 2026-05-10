# upload — 文件上传服务 (COS / OSS)

基于 Gin + Viper 的轻量上传服务，支持腾讯云 COS 和阿里云 OSS，通过 `APP_ENV` 环境变量切换。

## 目录

```
upload/
├── main.go           # 入口: APP_ENV=COS|OSS
├── cos_client.go     # COS HMAC-SHA1 签名上传
├── oss_client.go     # OSS HMAC-SHA1 签名上传
├── handler.go        # POST /v1/upload/file
├── config.go         # Viper 配置
├── go.mod / go.sum
├── Dockerfile
├── README.md
└── k8s/
    ├── configmap.yaml    # 生产云凭证
    ├── deploy.yaml       # 生产 Deployment + Service + Route
    └── test/             # 测试环境
```

## 本地开发

```bash
go mod tidy

# COS 模式
go run .

# OSS 模式
APP_ENV=OSS go run .
```

## 构建

```bash
go mod tidy

# 本地编译
go build -ldflags="-s -w" -o upload .

# Docker (生产)
podman build -t hub.jsesp.cn:5000/library/upload:prod .

# Docker (测试)
podman build -t hub.jsesp.cn:5000/library/upload:test .
```

## 部署

手动部署，不走 Tekton。

```bash
# 1. 编辑 k8s/configmap.yaml 填入云凭证
#    cos.secret_id / cos.secret_key  → 腾讯云
#    oss.access_key_id / oss.access_key_secret → 阿里云

# 2. 生产环境
kubectl apply -f k8s/

# 3. 测试环境
kubectl apply -f k8s/test/
```

## 接口

### `POST /v1/upload/file`

```
Content-Type: multipart/form-data

字段:
  file  (binary)  要上传的文件
  key   (string)  对象路径, 可选, 默认 uploads/{timestamp}.ext

响应 200:
{
  "data": {
    "url":  "https://bucket.cos.ap-guangzhou.myqcloud.com/uploads/xxx.jpg",
    "key":  "uploads/xxx.jpg",
    "size": 12345
  }
}
```

### `GET /health`

```json
{"status": "ok", "service": "upload", "timestamp": "2026-05-11T..."}
```

## 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `APP_ENV` | `COS` | 后端选择: `COS` / `OSS` |
| `CONFIG_DIR` | `.` | 配置文件目录 (K8s 挂载到 `/config`) |
