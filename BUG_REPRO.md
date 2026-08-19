# BUG_REPRO

## Bug 是什么
`BootstrapSample` 与 `sampling_policy.go` 中的采样函数原地复用底层数组，污染输入切片。

## 如何触发
对已有切片执行采样后继续读取原切片。

## 错误信息
原切片被改写为 `[5 4 3 4 5]` 之类。
