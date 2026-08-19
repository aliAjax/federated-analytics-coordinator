# BUG_REPRO

## Bug 是什么
`Manager` 的参与者 map 读写无锁，并发调用 `AddParticipant` 与 `SupportedMetrics` 时触发 data race 和并发 map 读写崩溃。

## 如何触发
多个 goroutine 并发注册参与者并查询支持指标。

## 错误信息
```
WARNING: DATA RACE
... internal/federation/manager.(*Manager).AddParticipant()
... internal/federation/manager.(*Manager).SupportedMetrics()
```
