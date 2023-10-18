# serverlessbench2.0

Serverlessbench2.0是一款**一键式**、**可定制**的跨平台Serverless test benchmark。它预设了若干测试用例，能够方便地衡量云应用和平台的性能；同时也支持用户扩展新的用例、待测试应用，以及支持新的平台。

**注意：目前workflow功能尚在开发中，并未完善**

## Prerequisite
目前主要支持在Knative以及Openwhisk上运行。
* Kubernetes $\geq$ 1.8
* Knative(with kn)
  * 部署方式：需要安装kn
* Openwhisk(with wsk)
  * 部署方式：基于kind k8s部署

其他平台仅支持部分基础功能，暂不支持一键测试，详见代码内对应的接口实现。

## Quick start

1. 进入项目根目录，编译二进制命令行工具
   
    ```bash
    cd serverlessbench2
    make
    ```
    
2. 进行参数配置,  登录docker

    ```bash
    ./cli config --Provider <provider-name> --PythonPath <your-python-interpreter-path> --DockerUserName DockerUserName --TestResultDir ./testResult
    
    docker login -u xxx -p xxx
    ```

    **注意:provider-name目前支持knative和openwhisk, your-python-interpreter-path为python3的路径**

3. 部署并调用函数，得到返回结果
   
    ```bash
    ./cli create -d helloworld ./apps/function/helloworld
    ./cli invoke helloworld text serverlessbench2
    ```
    
    ```json
    {"text":"serverlessbench2:hello world!"}
    ```

4.  进行一键式测试！

    ```bash
    ./cli apply example/exectest.yaml
    ```

   在./testResult中查看测试结果。

## cli命令行工具

cli是serverlessbench2用于和平台交互的二进制工具，封装了各平台的异构性，且能够方便地扩展到多平台。目前，cli支持的操作如下：

1. 部署函数：`cli create <funcName> <srcFilePath1> <srcFilePath2> .... <requirementPath> ` `cli create -d <funcName> <funcDirPath> `
2. 删除函数：`cli delete <funcName>`
3. 列出函数：`cli list`
4. 调用函数：`cli invoke <funcName> <paramName1> <paramValue1> ....`

5. 进行一键式测试：`cli apply <testConfigFliePath>`

更详细的介绍见项目目录中`docs/cli`文件。

## 进行自己的测试

serverlessbench2的测试流程由配置文件控制。可以使用预设的测试用例，或者在自己实现的测试用例上工作。

### 填写配置文件

需要填写待测试应用、一个实例如下。各项的说明见`docs`.

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

之后使用`cli apply`命令即可运行。

### 扩展测试用例和应用

如何扩展测试用例见`docs/testcase`。

按照serverlessbench2.0规范编写的、采用如下统一入口点的应用都可以直接进行测试。目前只支持Python编写的应用。

```python
def handler (event, context={}):
   param = event.get('paramName')
  # do sth
   return {"paramName":"someValue"}
```

## 代码测试
更新代码后，请运行项目测试用例，以确保不引入错误。具体的，在项目根目录下使用`go test -v ./test/*`命令。用于测试的文件包括：
* `a_prequisite_test.go`: 测试是否登录Docker，是否设置配置文件，以及cli是否能正常创建。
* `b_cli_test.go`: 测试cli的基本功能是否正常，包括修改配置文件、函数和函数链的相关操作（创建、调用、删除）。
* `c_testcase_test.go`: 测试ServerlessBench2的各测试用例是否能正常工作。

**请不要修改function文件夹中的`strLen`和`numAddOne`子目录，它们被用于进行项目测试。**

## What's More?

在项目目录中的`docs`文件夹，对本项目的用法、代码等做了详细介绍。具体的：

* `docs/cli.md`:介绍了项目的整体架构、cli命令行工具相关内容，包括各接口的说明。
* `docs/platform.md`:介绍了cli各接口在各平台上的实现，以及如何扩展到新的云平台。
* `docs/testcase.md`:介绍了测试流程的代码逻辑，以及如何进行测试。
* `docs/custom.md`: 介绍了在Serverlessbench2.0中实现新应用、扩展用例、扩展到新的平台。

-----

# 未完成功能

## cli命令行工具Workflow部分

cli workflow功能尚未完成，具体如下：

1. 删除workflow：`cli delete -f <workflowName>`
2. 调用workflow：`cli invoke -f <workflowName> <paramName1> <paramValue1> ....`

更详细的介绍见项目目录中`docs/cli`文件。