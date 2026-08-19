# BUG_REPRO

## Bug 是什么
`CompactShares` 使用 `shares[:0]` 原地压缩，复用调用方底层数组，污染传入的 shares 切片。`CloneShares` 返回同一引用，`UniqueParticipants` 不去重，`MergeShares` 缺 taskID 校验，`ParticipantCount` 误计空参与者。

## 如何触发
调用 `Engine.MergeShares` 后继续使用传入的 shares 切片，或直接调用上述函数。

## 错误信息
原切片被改写为 `[p1 p2 p2]` 之类的重复/缺失内容。
