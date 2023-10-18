# 扩展新的云平台

需要实现cli interface规定的所有接口。（可在internal文件夹内实现）具体的：

在`internal/benchinterface/factory.go` 内注册新的云平台。

```go
func providerFactory() Provider {
	p := CheckProvider()
	switch p {
	case "knative":
		return ProviderKnative{}
	case "openwhisk":
		return ProviderOpenwhisk{}
	case "openfaas":
		return ProviderOpenfaas{}
	case "fission":
		return ProviderFission{}
	default:
		return nil
	}
}
```

并实现Provider的所有接口函数。

```go
type Provider interface {
	CreateFunction(function_name string, memory_size int, source_code_file ...string) error
	DeleteFunction(function_name string) error
	ListFunction() error
	InvokeFunction(function_name string, params ...string) error
}
```

# 扩展新的应用

最好在apps/function文件夹内实现新的应用。函数的入口点需要按照统一接口格式书写。并且，apps/inputs文件夹内可以添加对该函数的默认输入参数，方便测试时调用。

具体的，新函数的代码应该全部实现在一个文件夹内：
```
apps 
	- function
		-- your function
```
其拥有一个名为`index.py`的入口点, 其中含有一个handler函数，返回字典形式，作为其执行入口：
```python
# index.py

def handler (event, context={}):
   return {}
```
需要提供一个依赖文件路径（形如xxx/requirements.txt）。若应用不需要安装依赖，也应给其提供一个空文件的路径。

# 扩展新的测试用例

扩展新的测试用例主要有以下几个步骤：

1. 设置测试用例类别，并注册测试用例。
2. 选择或实现新的Tesecase Runner。
3. 实现该测试用例的Metric Code。
4. 定义测试流程

## 注册测试用例

测试用例的种类目前为single，可以自己扩展新的类型。具体的，在`templates/testcase.yaml`中注册：

```yaml
execution-time: 
  type: single
  dir: ./testcase/testcase1-Execution-Time
coldstart-time: 
  type: single
  dir: ./testcase/testcase2-ColdStart-Time
warmstart-time: 
  type: single
  dir: ./testcase/testcase3-WarmStart-Time
cost-performance: 
  type: single
  dir: ./testcase/testcase4-Cost-Performance
```

注册之后，`ApplyYaml`函数就能够在解析测试配置文件之后，获取测试用例的代码路径。

## Tesecase Runner

一类测试用例对应一类Tesecase Runner，目前Tesecase Runner为single。若测试用例的类型为这两种之一，则在测试时会自动调用相应Runner，无需额外设置。

若是新的类型，则需自己实现Tesecase Runner。具体的：

1. 分别在Runner定义处和工厂函数中注册新的Runner。

   ```go
   // internal/runner/testcase-runner.go
   
   package runner
   
   type SingleRunner struct {
   	TCRunner
   }
   
   // internal/runner/factory.go
   func RunnerFactory(appType string) TCRunner {
   	switch appType {
   	case "single":
   		return SingleRunner{}
   	}
   	return nil
   }
   ```

2. 在对应runner的子文件中，实现`TestcaseRunner`函数。例子可参照`singleEntry`。大致的流程是：
   	- 通过`config`包中得到函数，获取python解释器路径等参数
   	- 使用当前Testcase文件夹中的Metric code包装原应用代码。
   	- 调用测试脚本。

## Metric Code

Metric Code有类似的格式：

1. 以`index.py`命名，作为函数新的入口点。
2. 调用`indexUserFunc`函数（被重命名的原函数入口点）进行处理。
3. 在外层进行时间戳收集等操作。

参照已给出的测试用例的Metric Code，可以定义自己想要收集的Metric Code。

## 定义测试流程

测试流程在测试用例文件夹中的`test.py`文件定义。都遵循类似的流程：

1. 通过`testcase/util/argparse.py`中的`argparse`类获取参数。
2. 通过`testcase/util/cli_scripts.py`中的control code和平台交互，预热、刷新函数等，获取Metric Code。
3. 处理Metric并输出结果。

argparse能够获取的参数在其对应文件中。扩展时可按需使用。

cli_scripts中的函数需要固定的传参格式，可根据其他测试用例加以模仿。

-----

# 未完成功能

## 扩展新的云平台的workflow接口
在实现上文的cli interface规定的接口外，还需实现如下接口：

```go
type Provider interface {
	CreateWorkflow(flow_name string, func_info ...map[string](map[string]string)) error
	DeleteWorkflow(flow_name string) error
	InvokeWorkflow(function_name string, params ...string) error
}
```

## 扩展新的workflow测试用例

扩展新的测试用例主要有以下几个步骤：

1. 设置测试用例类别，并注册测试用例。
2. 选择或实现新的Tesecase Runner。
3. 实现该测试用例的Metric Code。
4. 定义测试流程

### 注册测试用例

若要添加workflow类的测试用例。需要在`templates/testcase.yaml`中注册：

```yaml
communication-latancy: 
  type: workflow
  dir: ./testcase/testcase5-Communication-Latancy
varied-resource-requirements: 
  type: workflow
  dir: ./testcase/testcase6-Varied-Resource-Requirements
```

注册之后，`ApplyYaml`函数就能够在解析测试配置文件之后，获取测试用例的代码路径。

### Tesecase Runner

一类测试用例对应一类Tesecase Runner，故目前Tesecase Runner也有workflow类型。