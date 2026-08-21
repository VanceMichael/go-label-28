# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

采购员同时向两家供应商询饲料价格，一家很快报错后接口已经返回，但另一家请求仍长期挂在后台，没有收到这批询价已结束的通知；两家都成功时，报价列表还会因为谁先响应而颠倒，后续配方把价格对应错原料。请修复并行询价的生命周期和结果归位：首个失败要终止并收齐同批工作，成功结果必须保持采购单中的原料顺序，并通过竞态检查。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-28
- 仓库地址：https://github.com/VanceMichael/go-label-28.git
- parent SHA：d402293fe16c5c30bb8d723bb4550d3dd4892e8c

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-28.git bug-repro
cd bug-repro
git checkout --detach d402293fe16c5c30bb8d723bb4550d3dd4892e8c
go test ./internal/nutrition -run ^TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/nutrition -run ^TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder$ -count=1
--- FAIL: TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder (0.21s)
    --- FAIL: TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder/provider_failure_cancels_sibling (0.21s)
        quotes_test.go:37: sibling quote did not observe cancellation
    --- FAIL: TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder/results_retain_request_order (0.00s)
        quotes_test.go:56: quotes = [{IngredientCode:soy Supplier:primary UnitPrice:10} {IngredientCode:corn Supplier:primary UnitPrice:10}]
FAIL
FAIL	go-base/internal/nutrition	0.239s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/nutrition -run ^TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder$ -count=1
--- FAIL: TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder (0.20s)
    --- FAIL: TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder/provider_failure_cancels_sibling (0.20s)
        quotes_test.go:37: sibling quote did not observe cancellation
    --- FAIL: TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder/results_retain_request_order (0.00s)
        quotes_test.go:56: quotes = [{IngredientCode:soy Supplier:primary UnitPrice:10} {IngredientCode:corn Supplier:primary UnitPrice:10}]
FAIL
FAIL	go-base/internal/nutrition	0.205s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

soy 供应商返回指定错误时，CollectQuotes 必须保留该错误并取消、等待 corn 的同批调用结束，不能遗留后台询价；两家都成功且响应顺序相反时，结果仍按 corn、soy 的请求顺序归位。TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder 在 -race 下要由两个子场景失败转为全部通过，nutrition 原有用例、其他包回归和 go build ./... 保持绿色，不得串行化所有供应商、按完成时间改写采购顺序或削弱取消观察断言。
