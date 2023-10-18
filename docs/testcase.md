# 测试用例相关

目前共支持4个测试用例：

* 针对单个函数：
  * 执行时延测试：多次运行函数，收集执行时延并给出分位点。
  * 冷启动时延测试：多次调用函数，每次调用后刷新到冷启动状态，测试从发出请求到实际运行的冷启动时间。
  * 热启动时延测试：先预热后多次调用，测试热启动时间。
  * 性价比测试：给函数设置不同的最大内存，测试其运行时间，用时延*内存大小衡量价格。

## 测试流程代码说明

一次测试由`cli apply test.yaml`完成。经过的具体流程有：

1. cmd apply读取yaml文件，转换为预定义的`util.TestYaml`结构体形式，交由`internal/benchinterface/interface.ApplyYaml()`处理。
2. `ApplyYaml(）`解析结构体，转换为存有函数、workflow信息的map，遍历test metric和待测试app，根据测试类型交给相应的runner。
3. runner读取函数路径、结果路径等参数，包装用户代码，调用python测试脚本。
4. python测试脚本根据需要进行函数的部署、刷新、调用等操作，将结果存储进指定的路径。

### ApplyYaml() 统一接口解析

`internal/benchinterface/interface.ApplyYaml()`相当于控制平面，解析每次测试的参数，并依次将配置文件中所有testcase在每个待测试应用上运行。

它将结构体中`Function` 字段的所有函数解析进一个map，格式如下：

```
{func_name: {src:源文件路径， req:依赖路径， memory:内存大小}
```

之后，它分别遍历所有的metrics和apps，根据app类型实例化对应的runner，并转交其处理. testcase也有类型(single/workflow). 遍历时,每个testcase只在和其类型相同的app上测试：

```go
TCrunner := runner.RunnerFactory(appType)
if err := TCentry.TestcaseRunner(testName, testcaseDir, appName, func_path[appName], testApp.Param); err != nil {
	return err
}
```

### 测试的Runner

Runner主要的工作是根据testcase要求，包装用户函数的进入点，从而收集指标。它们有不同的包装用户代码的方式。和cli接口类似，也采用了工厂模式生成对应的runner。如下，根据应用类型生成：

```go
serverlessbench2/internal/runner/factory.go

package runner

type TCrunner interface {
	TestcaseRunner(testName string, testcaseDir string, appName string, appInfo interface{}, param string) error
}

func RunnerFactory(appType string) TCrunner {
	switch appType {
	case "single":
		return SingleRunner{}
	}
	return nil
}

```

#### 单函数Runner

针对单个函数的Runner(serverlessbench2/internal/runner/singleRunner.go)会将待测试函数的代码目录中的原进入点(index.py)替换为testcase的index.py，且将原来的index.py重命名，调用原来的进入点以运行函数。之后，它会以终端命令的形式启动Python测试脚本。

以Testcase1为例，这是其index.py:

 ```python
import indexUserFunc
import time

def getMillis():
    return round(time.time() * 1000)

def handler (event, context={}):
   startTime = getMillis()
   indexUserFunc.handler(event, context)
   returnTime = getMillis()
   return  {'startTime': startTime, 'returnTime': returnTime}

 ```

indexUserFunc为原先的index.py。

为存储结果，Runner会在配置文件的结果存储目录下新建以待进行testcase为名的文件夹（若已有则跳过），并将它设置为最终的结果存储路径：

```go
	_, err := os.Stat(testResultDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(testResultDir, 0777)
			if err != nil {
				return err
			}

		}
	}
```

## testcase代码说明

各Testcase的结构大体上是相似的。如下。

* index.py
* test.py

`index.py`用于替换掉用户代码中的进入点，帮助收集指标。以testcase1为例，在其handler中，获取了调用用户函数前、后的两个时间戳，从而可以计算出用户代码的执行时间。

若是对Workflow的测试用例，其`index.py`会使用相同的参数名传递链中每个函数的Metric，以便最后收集。

`test.py` 是主要的测试文件，进行函数的调用、收集结果等操作。结果存储路径由Runner给定。所需的参数由命令行解析得到：

```python
serverlessbench2/testcase/args/argparse.py

def init_args(params):
    arg_parser = argparse.ArgumentParser()
    arg_parser = add_argument_base(arg_parser)
    args = arg_parser.parse_args(params)
    return args

def add_argument_base(arg_parser):
    arg_parser.add_argument('--srcPath', default='', help='')
    arg_parser.add_argument('--memory', type=int, default=128, help='')
    arg_parser.add_argument('--param', default={}, type=ast.literal_eval, help='')
    arg_parser.add_argument('--appName',  default="", help='')
    arg_parser.add_argument('--cliBase',  default="", help='')
    arg_parser.add_argument('--testCaseDir',  default="",  help='')
    arg_parser.add_argument('--provider',  default="", help='')
    arg_parser.add_argument('--resultDir',  default="", help='')
    return arg_parser
```

### Control code

Control Code时抽象出来的一组和平台交互的操作，从而使得testcase能够操作函数，进行调用，获取metric。每一种操作对应一个函数，存放于`serverlessbench2/testcase/util/cli_scripts.py`, 可供所有的测试用例使用。当前实现的control code函数如下：

1. `flush`.  刷新函数或者workflow，通过isFlow参数设置。设置needCold参数可以使刷新后的函数处于冷/热启动状态。目前每个平台采用各自的刷新策略，扩展新平台时需要修改。
2. `invoke`: 调用函数或者workflow。除了函数自身的返回值外，会连调用请求时间点、调用完成时间点一并包装返回，方便启动时间等的计算。
3. `preWarm`：创建子进程不断调用函数指导得到结果，从而完成预热。

## 使用方法说明

完成一次测试的大致流程如下：

1. 进入serverlessbench2项目根目录，`make`生成cli 二进制文件。

2. 通过`cli config`配置云平台种类、Docker用户名、结果存储路径、Python解释器路径。
3. 创建并填写YOURNAME.yaml测试配置文件，也将其放置于根目录。
4. `cli apply YOURNAME.yaml`执行测试。

### 测试配置文件的填写

测试配置文件的格式如下：

```yaml
name: 
metric: 
  default:
    - name: 
  custom:
platform: 
resultpath: 
component:
  function:
    - name: 
      dirpath: 
      reqpath: 
      memory: 
     
test:
  - name: 
    type: 
    param: 

```

各个域的说明如下：

* name: 该测试的名称
* metric：default下的name列表填写serverlessbench2自带的metric名称（若写多个，则都会执行）。custum为用户自定义的testcase（TODO）
* platform：测试基于的云平台名称
* resultpath：结果存储路径
* component：function中填写本次测试中可能用到的所有函数信息；workflow时可能用到的所有workflow信息。
* test：实际执行测试的应用。可以是component中的单个函数或是workflow。需在type域中指明类型（single/workflow），在param域中填写调用参数（json字符串）。可以使用自带默认参数，此时参数Default域内需指明参数路径（在inputs文件夹内）

一个示例如下：

```yaml
name: test-ffmpeg
metric-controller: 
  default:
    - name: execution-time
  custom:
platform: knative
resultpath: ./testResult
component:
  function:
    - name: prime
      dirpath: ./function/getprime
      reqpath: ./function/requirements.txt
      memory : 128
    - name: len
      dirpath: ./function/strlen
      reqpath: ./function/requirements.txt
      memory : 128
    - name: httprequest
      dirpath: ./function/httpRequest
      reqpath: ./function/requirements.txt
      memory : 128
    - name: uploader
      dirpath: ./function/uploader
      reqpath: ./function/requirements.txt
      memory : 128
    - name: ffmpeg
      dirpath: ./function/ffmpeg
      reqpath: ./function/requirements.txt
      memory : 128

test:
  - name: ffmpeg
    type: single
    param: 
      default: 
        - name: default1.txt
        - name: default2.txt
        - name: default3.txt
        - name: default4.txt
        - name: default5.txt
      other: 
         - value: "{\"payload\":'\\\'{\\\\\\\"msg\\\\\\\":\\\\\\\"youaregood\\\\\\\"}\\\''}"

```

-----

# 未完成功能

# Workflow测试用例相关

Workflow有2个测试用例：

* 针对函数链
  * 通信延迟测试：测试函数链中各函数的启动时间差，作为通信时延。
  * 函数资源需求测试：分别衡量运行整个函数链和分开运行链中各函数的运行时延和价格。

## Workflow测试流程代码说明

Workflow流程与single function基本一致。

### ApplyYaml() 统一接口解析

每个workflow解析信息如下：

```
{flow_name:[{step1_name: {src:源文件路径， req:依赖路径， memory:内存大小}},...]}
```

### 测试的Workflow Runner

与single function基本一致：

```go
serverlessbench2/internal/runner/factory.go

package runner

type TCrunner interface {
	TestcaseRunner(testName string, testcaseDir string, appName string, appInfo interface{}, param string) error
}

func RunnerFactory(appType string) TCrunner {
	switch appType {
	case "single":
		return SingleRunner{}
	case "workflow":
		return WorkflowRunner{}
	}
	return nil
}

```

#### Workflow Entrypoint

Workflow的runner和single版本的大体流程相似，但略有不同。具体地：

1. 给测试脚本转入参数时，传入workflow中每个函数的源文件夹路径、内存大小等参数的列表。
2. 包装函数时，包装workflow中每个step函数。

## Workflow testcase代码说明

Workflow Testcase的结构与single function一致。大体上是相似的。

其中`test.py`在获取参数时会解析传入的step列表：

```go
   srcPath = args.srcPathList
   reqPath = args.reqPathList
   memList = args.memSizeList
   nameList = args.stageNameList
   param = args.param
   appName = args.appName
   testCaseDir = args.testCaseDir
   provider = args.provider
   resultDir = args.resultDir

   stageNames = nameList.split(',')
   stagePaths = srcPath.split(',')
```

同时，调用workflow的对应接口进行刷新、预热等操作。

### Workflow测试配置文件的填写

测试配置文件的格式如下：

```yaml
name: 
metric: 
  default:
    - name: 
  custom:
platform: 
resultpath: 
component:
  function:
    - name: 
      dirpath: 
      reqpath: 
      memory : 
  workflow:
    - name: 
      stage:
        - funcname: 
     
test:
  - name: 
    type: 
    param: 

```

各个域的与Single function一致。

一个示例如下：

```yaml
name: test-ffmpeg
metric-controller: 
  default:
    - name: execution-time
  custom:
platform: knative
resultpath: ./testResult
component:
  function:
    - name: prime
      dirpath: ./function/getprime
      reqpath: ./function/requirements.txt
      memory : 128
    - name: len
      dirpath: ./function/strlen
      reqpath: ./function/requirements.txt
      memory : 128
    - name: httprequest
      dirpath: ./function/httpRequest
      reqpath: ./function/requirements.txt
      memory : 128
    - name: uploader
      dirpath: ./function/uploader
      reqpath: ./function/requirements.txt
      memory : 128
    - name: ffmpeg
      dirpath: ./function/ffmpeg
      reqpath: ./function/requirements.txt
      memory : 128

  workflow:
    - name: seq
      dirpath: ./workflow/seq
      stage:
        - funcname: len
        - funcname: prime
test:
  - name: ffmpeg
    type: single
    param: 
      default: 
        - name: default1.txt
        - name: default2.txt
        - name: default3.txt
        - name: default4.txt
        - name: default5.txt
      other: 
         - value: "{\"payload\":'\\\'{\\\\\\\"msg\\\\\\\":\\\\\\\"youaregood\\\\\\\"}\\\''}"

```