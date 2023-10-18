# 整体架构介绍

项目的文件结构如下：

* cmd
  * cli
  * cmd
* apps
* example
* docs
* internal
  * benchinterface
  * config
  * entrypoint
  * util
* templates
* test
* testcase
* testResult
* Makefile

cmd目录用于生成cli命令行工具，apps目录存放所有的被测试函数以及所提供的默认输入参数（如果有）、internal目录是**核心**目录，有各命令在各平台上的具体实现，以及测试的entrypoint。templates中存放在各平台上创建、删除函数等操作可能用到的模板文件（如Dockerfile）。testResult是默认的结果存放路径。test文件夹中存放用于测试项目的代码。

具体使用时，create、delete等操作由编译出的cli二进制文件提供，使用cli apply命令即可根据用户提供的配置文件执行测试。
## cmd说明及cli的接口

这是项目编译的入口，同时也规定生成的cli的对外接口。目前包括有以下接口：

* **cli apply**:  爬取测试yaml文件，提取所需函数、需要进行的测试等信息，并进行测试。
* **cli create**: 创建函数，可从一列函数文件路径创建函数，或是遍历某一个文件夹。
* **cli delete**：删除函数。
* **cli list**：列出当前云平台上所部署的所有函数。
* **cli config**：写入或读取配置文件。
* **cli invoke**: 调用函数。

每种函数的逻辑都是类似的：做对参数的基本判断和处理，然后转交给`serverlessbench2/internal/benchinterface`中的对应接口作处理。

### cli apply

**使用格式**：`cli apply xxx/yamlname.yaml`

该命令通过传入的配置文件路径读取测试用yaml，解析后转交给`benchinterface.ApplyYaml()`，执行测试。

测试的yaml格式定义在`serverlessbench2/internal/util/yaml.go`中。

### cli create

**使用格式**：`cli create funcname sourcepath1 sourcepath2 .... requirementpath` 或 `cli create -d funcname funcsourcedir `

**flag**:  `--memory/-m`, 指定函数的最大内存大小。

该命令读取函数各文件路径，或是从某文件夹中爬取文件（此时依赖文件也需包括在文件夹内），然后创建函数。

### cli list

**使用格式**：`cli list` 或

能够列出平台上部署的所有函数和workflow。

### cli delete

**使用格式**：`cli delete funcname `

删除函数。

### cli config

**使用格式**：`cli config --configname value...`

修改配置文件。所有需设置的配置项如下：

*   **Provider** : 当前云平台名称。值可设置为：
    - **knative**
    - **openwhisk**
    - **openfaas（待后续开发）**
    - **fission（待后续开发）**

*   **PythonPath**：Python解释器路径。
*   **DockerUserName**：存放函数镜像的Docker用户仓库的用户名。
*   **TestResultDir**：测试结果存储路径。

### cli invoke

**使用格式**：`cli invoke funcname param1 value1 ....`

调用函数。传入的参数都被解析成`string` 格式，需要在用户函数中重新转换。

-----

## 未完成功能

### workflow delete

**使用格式**：`cli delete -f workflowname`

删除workflow。

### workflow invoke

**使用格式**：`cli invoke -f flowname param1 value1.... `

调用workflow。传入的参数都被解析成`string` 格式，需要在用户函数中重新转换。