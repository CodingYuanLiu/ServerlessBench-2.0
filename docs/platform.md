# 各平台接口实现说明

每个平台上所需实现的接口种类、传参格式都是相同的，只是内部实现的具体细节不同。故采用工厂模式完成各平台上的实现。具体地，将所需实现的所有接口抽象成一个接口集合：

```go
serverlessbench2/internal/benchinterface/factory.go

type Provider interface {
	CreateFunction(function_name string, memory_size int, source_code_file ...string) error
	DeleteFunction(function_name string) error
	InvokeFunction(function_name string, params ...string) error}
```

每一个平台对应该抽象结构体的一个实例：

```go
serverlessbench2/internal/benchinterface/providers.go

type ProviderKnative struct {
	Provider
}

type ProviderOpenwhisk struct {
	Provider
}

type ProviderOpenfaas struct {
	Provider
}

type ProviderFission struct {
	Provider
}
```

调用时（例如cli create），首先会到达统一接口处，然后根据配置文件构造对应于当前平台的provider，调用它相应的处理方法。create的统一接口如下：

```go
serverlessbench2/internal/benchinterface/interface.go 135

func CreateFunction(function_name string, memory_size int, source_code_file ...string) error {
	provider := CheckProvider()
	if provider == "" {
		return nil
	}
	p := providerFactory()
	err := p.CreateFunction(function_name, memory_size, source_code_file...)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
```

每一个平台所对应的结构体都需实现上述所有方法。若要支持新的平台，只需新增相应实例结构体，在`internal/benchinterface/factory.providerFactory()`中注册，然后实现其所有方法即可。

接下来介绍各平台上的接口实现细节。

**注**：cli需要用到Templates中的模板文件。为了解决其路径依赖问题，采用了go-bindata包将其二进制打包。获取文件内容时采用`util.LoadFile(filename, targetDir)`函数可将文件加载至指定目录，或用`util.Asset("resources/filename")`获取文件的二进制内容。

### Openwhisk

#### CreateFunction

首先获取 `templates/openwhisk/__main__.py`的文件内容，它是Openwhisk函数的entrypoint，用它来包装用户函数；之后在`/tmp`创建一个临时文件夹，来存放生成函数镜像所需的文件。将`__main__.py`、各源文件压缩，拷贝进临时文件夹，之后运行docker命令，将其挂载进容器并初始化环境，然后使用`wsk`命令创建函数。

#### DeleteFunction

调用wsk命令删除。

#### ListFunction

调用wsk list。

#### InvokeFunction

调用wsk invoke，将传入的参数加上-p flag拼接在命令后。

### Knative

#### CreateFunction

和openwhisk类似创建一个临时文件夹，加载knative template中的Dockerfile、server.py、service.yaml到临时文件夹。之后替换Docekrflie中用户名、service.yaml中的镜像名、函数名，等字段，创建并push函数镜像，然后使用kubectl apply部署服务。

#### DeleteFunction

使用kn命令行工具删除。

#### ListFunction

使用kn list，匹配其中的url地址，加工得到所有函数名。对workflow，改为查找broker的名称。

#### InvokeFunction

将参数组装成json格式，使用HTTP GET向函数的url发送请求。

### Openfaas（待后续开发）

#### CreateFunction

拷贝模板文件夹压缩包到FaasTempDir并解压，创建新函数。之后使用用户的handler覆盖模板中的handler、添加依赖搜索路径。拷贝源文件并修改yaml配置后，使用`faas-cli up`部署函数。

#### DeleteFunction

使用faas-cli remove。

#### ListFunction

使用faas-cli list。

#### InvokeFunction

通过faas-cli describe获取函数URL，之后通过curl调用。

### Fission

和Openfaas类似。使用fission fn相关API。

-----

# 未完成功能

## 各平台workflow接口实现说明

```go
serverlessbench2/internal/benchinterface/factory.go

type Provider interface {
	CreateWorkflow(flow_name string, func_info ...map[string](map[string]string)) error
	DeleteWorkflow(flow_name string) error
	InvokeWorkflow(function_name string, params ...string) error
}
```

调用workflow时（例如cli create）与调用function相同，不过接口中的`CreateWorkflow`较为特殊：它没有在cmd中出现。实际上其由于参数格式原因，难以在命令行传参，故不对用户开放。他的`func_info ...map[string](map[string]string`是一列map，其中每一个对应着workflow中一个step的函数信息。 一个例子如下：

```
func_info : {name: {src:源文件路径， req:依赖路径， memory:内存大小}....
```

### Openwhisk workflow

#### CreateWorkflow

使用wsk命令基于已有的函数创建workflow。（所有函数必须事先部署）


#### DeleteWorkflow, InvokeWorkflow

同function。

### Knative workflow

#### CreateWorkflow

Knative的workflow需要broker、trigger组件来传递event（即调用命令和参数）。由于其名称限制，首先对workflow的名称作检查，使其不能包含某些特殊字符。之后加载`workflow.yaml`（用于部署函数sequence）、`invokeComponents.yaml`（用于部署broker、trigger、displayer）。

由于knative的sequence只支持go，我们需要对每一个step的代码作包装，即用sequence上的go函数组件来调用它。该wrapper也在templates中。具体地，遍历func_info，拷贝其源文件目录中所有文件、依赖文件，wrapper，以及部署函数所需的service.yaml等文件到KnTempDir，然后同上类似创建每一个step service。

之后，修改`workflow.yaml`、`invokeComponents.yaml`中的通配符，然后创建sequence和调用组件。

#### DeleteWorkflow

删除所部署的sequence，以及各调用组件。

#### InvokeWorkflow

首先将参数组装为json格式。调用组件中包括了一个配置有curl的POD，通过k8s的RESTClient() API在该容器中向所调用的workflow发送请求。最终结果会在displayer的日志中显示，获取并筛选信息，得到最后一次调用的结果。