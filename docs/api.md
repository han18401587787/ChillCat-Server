# ChillCat API 文档 v1

> 基础地址：`https://api.chillcatgo.com/api/v1`（本地开发：`http://localhost:8080/api/v1`）

---

## 通用说明

### 认证方式

```
Authorization: Bearer <token>
```

### 统一响应格式

```json
{
    "code": 0,
    "message": "success",
    "data": { ... }
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 10001 | 请求参数错误 |
| 10002 | 登录已过期 |
| 10003 | 无权限访问 |
| 10004 | 资源不存在 |
| 10006 | 服务器内部错误 |
| 20001 | 用户不存在 |
| 20002 | 用户已存在 |
| 20003 | 密码错误 |
| 30001 | 会员信息不存在 |
| 30002 | 会员已过期 |
| 30003 | 购买失败 |

---

## 1. 健康检查

### GET /health

```
Response 200:
{
    "code": 0,
    "message": "success",
    "data": {
        "status": "ok",
        "version": "1.0.0"
    }
}
```

---

## 2. 认证接口

### 2.1 注册

```
POST /api/v1/auth/register

Request:
{
    "username": "testuser",       // 4-32位，字母数字下划线
    "email": "test@example.com",
    "password": "123456",         // 6-64位
    "nickname": "测试用户"         // 可选，不传则用 username
}

Response 200:
{
    "code": 0,
    "message": "success",
    "data": {
        "user_id": 1,
        "username": "testuser",
        "token": "eyJhbGciOiJIUzI1NiIs..."
    }
}
```

### 2.2 登录

```
POST /api/v1/auth/login

Request:
{
    "username": "testuser",
    "password": "123456"
}

Response 200:
{
    "code": 0,
    "message": "success",
    "data": {
        "user_id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "avatar": "",
        "token": "eyJhbGciOiJIUzI1NiIs..."
    }
}
```

---

## 3. 用户接口（需认证）

### 3.1 获取用户信息

```
GET /api/v1/user/profile

Response 200:
{
    "code": 0,
    "message": "success",
    "data": {
        "user_id": 1,
        "username": "testuser",
        "email": "test@example.com",
        "nickname": "测试用户",
        "avatar": "",
        "status": 1
    }
}
```

---

## 4. 会员接口（需认证）

### 4.1 获取会员信息

```
GET /api/v1/member/info

Response 200（会员）:
{
    "code": 0,
    "message": "success",
    "data": {
        "member_type": "yearly",
        "status": "active",
        "start_date": "2025-07-16T00:00:00Z",
        "end_date": "2026-07-16T00:00:00Z",
        "auto_renew": true,
        "is_valid": true,
        "remaining_days": 365
    }
}

Response 200（非会员）:
{
    "code": 0,
    "message": "success",
    "data": {
        "member_type": "",
        "status": "none",
        "is_valid": false,
        "remaining_days": 0
    }
}
```

### 4.2 获取商品列表

```
GET /api/v1/member/products

Response 200:
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": "monthly_001",
            "member_type": "monthly",
            "display_name": "月度会员",
            "price": 15.00,
            "duration_days": 30
        },
        {
            "id": "quarterly_001",
            "member_type": "quarterly",
            "display_name": "季度会员",
            "price": 40.00,
            "original_price": 45.00,
            "discount_tag": "省 11%",
            "duration_days": 90
        },
        {
            "id": "yearly_001",
            "member_type": "yearly",
            "display_name": "年度会员",
            "price": 128.00,
            "original_price": 180.00,
            "discount_tag": "省 29%",
            "duration_days": 365
        },
        {
            "id": "permanent_001",
            "member_type": "permanent",
            "display_name": "永久会员",
            "price": 399.00,
            "discount_tag": "一次购买，永久有效",
            "duration_days": -1
        }
    ]
}
```

### 4.3 获取权益列表

```
GET /api/v1/member/privileges

Response 200:
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": "unlimited_play",
            "title": "无限畅听",
            "description": "畅享全曲库，无限制播放",
            "icon": "music_note",
            "is_highlight": true,
            "available_for": ["monthly", "quarterly", "yearly", "permanent"]
        },
        {
            "id": "high_quality",
            "title": "无损音质",
            "description": "享受 Hi-Res 无损音质",
            "icon": "high_quality",
            "is_highlight": true,
            "available_for": ["quarterly", "yearly", "permanent"]
        },
        {
            "id": "download",
            "title": "离线下载",
            "description": "歌曲离线下载，无网络也能听",
            "icon": "download",
            "is_highlight": false,
            "available_for": ["monthly", "quarterly", "yearly", "permanent"]
        },
        {
            "id": "no_ad",
            "title": "去广告",
            "description": "免广告打扰，纯净体验",
            "icon": "block",
            "is_highlight": false,
            "available_for": ["monthly", "quarterly", "yearly", "permanent"]
        },
        {
            "id": "exclusive_content",
            "title": "专属内容",
            "description": "会员专属歌曲、歌单、直播回放",
            "icon": "star",
            "is_highlight": false,
            "available_for": ["yearly", "permanent"]
        }
    ]
}
```

### 4.4 购买会员

```
POST /api/v1/member/purchase

Request:
{
    "product_id": "yearly_001"
}

Response 200:
{
    "code": 0,
    "message": "success",
    "data": {
        "order_no": "CC17211234560001",
        "member_info": {
            "member_type": "yearly",
            "status": "active",
            "start_date": "2025-07-16T00:00:00Z",
            "end_date": "2026-07-16T00:00:00Z",
            "auto_renew": true,
            "is_valid": true,
            "remaining_days": 365
        }
    }
}
```

### 4.5 购买历史

```
GET /api/v1/member/history?page=1&page_size=10

Response 200:
{
    "code": 0,
    "message": "success",
    "data": {
        "list": [
            {
                "order_no": "CC17211234560001",
                "member_type": "yearly",
                "order_status": "success",
                "amount": 128.00,
                "payment_method": "mock",
                "created_at": "2025-07-16T10:00:00Z"
            }
        ],
        "total": 1,
        "page": 1,
        "page_size": 10
    }
}
```
