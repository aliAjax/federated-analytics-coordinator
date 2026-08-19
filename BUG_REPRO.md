# BUG_REPRO

## Bug 是什么
状态机新增 `retrying` 中间态后，漏配 `retrying -> collecting` 返回边，活跃状态判定和 replay 查询也未纳入 retrying。

## 如何触发
将任务置为 retrying 后尝试重新收集，并查询最新活跃状态。

## 错误信息
`invalid state transition from retrying to collecting`。
