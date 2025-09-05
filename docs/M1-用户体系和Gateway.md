# M1: 用户体系和 Gateway

**状态**: 🔄 进行中  
**开始时间**: 2025 年 9 月 5 日  
**预计完成**: 2025 年 9 月 6 日  
**当前进度**: 0%

## 目标

基于 M0 完成的 Auth Service，开发完整的用户管理体系和高性能 Gateway 服务，支持长连接和消息路由。

## 核心目标

1. **User Service (Go)**: 用户信息管理、好友关系、用户状态
2. **Gateway Service (C++)**: 长连接管理、消息路由、负载均衡
3. **服务间通信**: 完善 gRPC 通信和服务发现
4. **数据一致性**: 跨服务的数据同步和缓存策略

## 任务清单

### 1. User Service 设计和开发 🔄

**当前进度**: 0%

#### 数据模型设计

- [ ] 用户信息扩展模型
  - 昵称、头像、个性签名
  - 在线状态、最后上线时间
  - 隐私设置（是否允许被搜索等）
- [ ] 好友关系模型
  - 好友申请、同意/拒绝流程
  - 好友分组管理
  - 黑名单功能
- [ ] 用户设置模型
  - 消息推送设置
  - 隐私权限设置

#### API 接口设计

- [ ] 用户信息管理 API
  - 获取/更新用户资料
  - 上传头像
  - 修改状态
- [ ] 好友管理 API
  - 搜索用户
  - 发送/处理好友请求
  - 好友列表管理
  - 删除好友
- [ ] 设置管理 API
  - 隐私设置
  - 推送设置

#### gRPC 服务定义

- [ ] user.proto 定义
  - UserInfo, FriendRequest, UserSettings 消息
  - UserService gRPC 服务接口
- [ ] 与 Auth Service 的集成
  - 用户注册后自动创建用户档案
  - 登录时同步用户状态

#### 实现细节

- [ ] 数据库设计和迁移
- [ ] Repository 层实现
- [ ] Service 层业务逻辑
- [ ] HTTP + gRPC 双协议支持
- [ ] 与 Redis 缓存集成
- [ ] 日志系统集成

### 2. Gateway Service 架构设计 ⚪

**当前进度**: 0%

#### 技术选型

- [ ] 网络库选择评估
  - Boost.Beast (HTTP/WebSocket)
  - uWebSockets (高性能 WebSocket)
  - 自研基于 epoll/kqueue
- [ ] 构建系统设计
  - CMake 项目结构
  - 依赖管理 (vcpkg/Conan)
  - 编译配置

#### 架构设计

- [ ] 连接管理设计
  - 长连接池管理
  - 心跳保活机制
  - 连接状态同步
- [ ] 消息路由设计
  - 用户会话路由表
  - 负载均衡策略
  - 故障转移机制
- [ ] 协议设计
  - WebSocket 消息格式
  - 二进制协议优化
  - 协议版本兼容

#### 性能优化

- [ ] 内存池管理
- [ ] 零拷贝消息传递
- [ ] 多线程模型设计
- [ ] 监控和指标收集

### 3. 服务集成和通信 ⚪

**当前进度**: 0%

#### 服务发现

- [ ] 简单配置文件方式
- [ ] 健康检查机制
- [ ] 服务注册与发现

#### 数据一致性

- [ ] 跨服务事务处理
- [ ] 缓存更新策略
- [ ] 数据同步机制

#### 监控和日志

- [ ] 统一日志格式
- [ ] 性能指标收集
- [ ] 分布式链路追踪

## 技术架构

### 服务通信流程

```
Client (WebSocket)
    ↓
Gateway Service (C++)
    ↓ (gRPC)
User Service (Go) ←→ Auth Service (Go)
    ↓
Database + Redis
```

### 数据库设计

#### 用户信息表 (users_profile)

```sql
CREATE TABLE users_profile (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    nickname VARCHAR(50) NOT NULL,
    avatar_url TEXT,
    signature TEXT,
    status VARCHAR(20) DEFAULT 'offline', -- online, offline, away, busy
    last_seen TIMESTAMP,
    allow_search BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### 好友关系表 (friendships)

```sql
CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    friend_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL, -- pending, accepted, blocked
    created_at TIMESTAMP DEFAULT NOW(),
    accepted_at TIMESTAMP,
    UNIQUE(user_id, friend_id)
);
```

## 风险评估

### 技术风险

- **C++ Gateway 复杂度**: 网络编程和内存管理复杂性
- **性能目标**: 单机支持 10K+ 并发连接
- **跨语言通信**: Go-C++ gRPC 通信稳定性

### 解决方案

- 分阶段开发，先实现基础功能
- 充分的单元测试和性能测试
- 参考成熟的开源项目架构

## 验收标准

### 功能验收

- [ ] User Service 完整 API 实现
- [ ] Gateway 支持 WebSocket 长连接
- [ ] 用户上线状态实时同步
- [ ] 好友关系完整流程

### 性能验收

- [ ] Gateway 支持 1000+ 并发连接
- [ ] API 响应时间 < 100ms
- [ ] 消息传递延迟 < 50ms

### 质量验收

- [ ] 90%+ 代码覆盖率
- [ ] 完整的 API 文档
- [ ] 部署和运维文档
