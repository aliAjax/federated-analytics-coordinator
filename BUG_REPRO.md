# BUG_REPRO

## Bug 是什么
`Scheduler.LeaseContext` 忽略 context 取消，`Runner.Run` 传入 `context.Background()`，已取消的 context 仍会租用并完成任务；`Scheduler.Has` 与 `Runner` 参数校验也错误。

## 如何触发
创建已取消的 context，调用 `Runner.Run` 并传入可租用任务。

## 错误信息
已取消 context 仍返回任务数大于 0。
