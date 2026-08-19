# BUG_REPRO

## Bug 是什么
存储层未包装 `ErrNotFound`，错误链断裂，HTTP 层无法用 `errors.Is` 识别 not found，缺失任务调用 start 返回 422。

## 如何触发
对不存在的任务调用 `POST /api/v1/tasks/{id}/start`。

## 错误信息
HTTP 状态码为 422（应为 404）。
