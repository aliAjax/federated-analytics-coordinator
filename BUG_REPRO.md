# BUG_REPRO

## Bug 是什么
`Accountant.Settle` 使用 defer 无条件提交台账，delta 非法时仍把预留记录标记为 committed；台账 `Commit`/`Release` 未包装 `ErrEntryNotFound`。

## 如何触发
用 delta=1 的非法 `LedgerEntry` 调用 `Accountant.Settle`。

## 错误信息
错误返回后台账条目仍为 committed；not found 错误无法被 `errors.Is` 识别。
