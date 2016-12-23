opsmind gopsutil 维护说明
=====

gopsutil 是 frok 自 github.com/shirou/gopsutil，以便于 opsmind 项目对其单独维护和修改。但**所有的修改请尽量以『可提交给上游』为原则**。


### 引用和更新流程

opsmind 项目若需引用该库，为了维持 gopsutil 内部的 import 路径不变，请使用以下方式添加到一个项目中：

```
> govendor fetch github.com/shirou/gopsutil::github.com/opsmind/gopsutil
```

### 修改流程（以 xuzhaokui 为例）

1. 从 `github.com/opsmind/gopsutil` fork 到自己账号下（即: github.com/xuzhaokui/gopsutil）
1. 在本地 `$GOPATH/src/opsmind/` 目录下: `git clone git@github.com:xuzhaokui/gopsutil`
2. 在 opsmind 项目内的 `vendor/github.com/shirou/gopsutil` 目录内直接修改（方便编译和调试）
3. 使用目录同步工具将修改后的目录同步到 `$GOPATH/src/code.opsmind.com/gopsutil` 中
4. 在 `$GOPATH/src/code.opsmind.com/gopsutil` 中将修改以 PR 方式提交给 `github.com/opsmind/gopsutil`
5. 在 opsmind 项目内更新 vendor，更新过程同上面`引用和更新流程`

其中第三步的目录同步，可以如下操作（以 dog 项目为例）：

```
> rsync -a --delete --exclude='.git/' --exclude='.gitignore' $GOPATH/src/code.opsmind.com/dog/vendor/github.com/shirou/gopsutil/ $GOPATH/src/code.opsmind.com/gopsutil
```
