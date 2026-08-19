# BUG_REPRO

## Bug 是什么
策略标签索引 `IndexTags` 在写入 `p.TagIndex[tag]` 时未初始化外层 map，导致首次构建带 required_tags 的策略索引时触发 nil map 写入 panic。同一策略包中 `TagCount` 与标签匹配逻辑也存在错误。

## 如何触发
创建包含 `required_tags` 的 `policy.Rule`，再执行 `NewEvaluator` 构建索引；或调用 `TagCount` 与 `Evaluate` 验证标签匹配。

## 错误信息
```
panic: assignment to entry in nil map
goroutine ... [running]:
... internal/policy/domain.(*PolicySet).IndexTags(...)
```
