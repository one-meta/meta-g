# Meta-G（Meta 代码生成器）

## 作用

根据 entgo 模板生成的`data_map.go`文件，通过 [生成器](https://github.com/Xuanwo/gg) 生成 controller，service，查询参数、api 路由、api 测试用例的代码

## 为什么会有这个项目

Meta 基础框架完成之后，新建 entgo schema 时，需要手动添加 Controller、Service、api 路由之类，于是通过 [gg](https://github.com/Xuanwo/gg) 项目进行代码生成

## Installation

go >= 1.17

`go install github.com/one-meta/meta-g@latest`

或 clone 项目后，`go install`

**使用**

在 meta 项目根目录使用

或者任意目录与`data_map.go`同级目录使用 `meta-g`

### 其他方式（不建议）

**测试**

可以直接复制 data_map.go 到项目目录，然后`go run main.go  -merge=true`，即可覆盖生成代码，生成的代码在 app 目录下；此时可以在 app 目录下查看代码，然后复制到实际项目中

**开发**

`make build`之后复制可执行文件到 Meta 项目的 generator 目录下，通过 air 进行热加载，此时 air 配置会自动运行生成器，代码会生成到各个对应目录下（不会覆盖已经生成的代码）

**注意**：开发时不要使用 `-merge=true`,否则会替换掉已修改的代码

## 更多信息

controller.go，生成控制器文件和 swagger 注解

service.go，生成服务文件

api.go，生成路由文件

api_test_case.go，生成 api 测试用例

query_param，生成额外的时间查询参数（时间范围查询）
