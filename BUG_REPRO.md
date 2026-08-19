# BUG_REPRO

## Bug 是什么
`CompleteTask` 使用 defer 保存任务，命名返回值被 `SaveTask` 覆盖，预算超额错误被吞掉，任务被错误置为 Completed；完成策略相关函数也错误放行非法预算。

## 如何触发
对 Revealing 且 `Budget.Spent > Budget.Epsilon` 的任务调用 `CompleteTask`。

## 错误信息
调用返回 nil，任务状态被保存为 completed。
