---
trigger: always_on
---

---
inclusion: always
---

# Go 项目工程规范

## 1. 技术栈
- Go 1.24+
- go-ethereum (以太坊客户端)
- GORM (数据库 ORM)
- SQLite (数据库)
- Zap (日志库)

## 2. 代码结构
- 包结构清晰：`config/`, `model/`, `database/`, `utils/`, `wallet/`, `monitor/`
- 每个包职责单一
- 避免循环依赖

## 3. 编码规范
- 使用小函数，单一职责
- 错误处理：返回 error，不要 panic
- 使用 context.Context 管理生命周期
- 适量添加日志和注释（中文）
- 避免硬编码和魔数，使用常量或配置
- 使用 defer 确保资源释放
- 避免每次都生成新的测试文件，尽量使用main函数，避免main函数冲突
- 问题的回答要通过编译，不要出现编译问题

## 4. 并发规范
- 使用 channel 进行 goroutine 通信
- 使用 sync.WaitGroup 等待 goroutine 完成
- 避免共享内存，优先使用 channel
- 使用 context 控制 goroutine 生命周期
- 使用 worker pool 模式处理高并发任务

## 5. 数据库规范
- 使用 GORM 进行数据库操作
- 所有查询使用参数化，防止 SQL 注入
- 使用事务处理关联操作
- 添加适当的索引（TxHash 唯一索引）

## 6. 性能优化
- WebSocket 订阅优先于轮询
- 使用异步处理，避免阻塞
- 批量操作使用 channel + worker pool
- 避免 N+1 查询

## 7. 错误处理
- 所有错误都要处理或记录
- 使用 zap.Error() 记录错误日志
- 数据库操作失败要记录详细信息

## 8. 沟通规范
- 技术讨论全程使用中文
- 问题回答聚焦不发散
- 避免每次回答都生成总结文档
- 修改现有文件，减少新建文件
