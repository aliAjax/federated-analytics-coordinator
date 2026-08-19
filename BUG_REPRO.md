# BUG_REPRO

## Bug 是什么
`MetricStore.Snapshot` 返回内部 map 引用，并发写入与快照遍历触发 data race；`Count`/`Delete` 未正确加锁，`ValidateMetricMap` 未校验负 Count。

## 如何触发
多个 goroutine 并发 Put 指标并遍历 Snapshot 返回的 map。

## 错误信息
`WARNING: DATA RACE`。
