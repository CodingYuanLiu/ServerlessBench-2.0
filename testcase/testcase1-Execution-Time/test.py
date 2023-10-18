import os
import sys
import time
import argparse
from urllib.parse import ParseResult

import sys

os.path.dirname(__file__)
sys.path.append(os.path.join(os.path.dirname(__file__), ".."))
from util import argparse
from util import cli_scripts

TEST_TIME = 60


def parseResult(rawResult):
    startTime = int(eval(rawResult[0])['startTime'])
    returnTime = int(eval(rawResult[0])['returnTime'])
    return startTime, returnTime


def statResult(latencies, failRequest, mode, file):
    latencies.sort()
    totalNum = len(latencies)
    if totalNum == 0:
        file.write("all request fail")
        return
    _p50Latency = latencies[int(totalNum * 0.5) - 1]
    _p75Latency = latencies[int(totalNum * 0.75) - 1]
    _p90Latency = latencies[int(totalNum * 0.9) - 1]
    _p99Latency = latencies[int(totalNum * 0.99) - 1]
    file.write("%s latency (ms):\n" % mode)
    file.write("50%\t75%\t90%\t99%\n")
    file.write("%d\t%d\t%d\t%d\n" %
               (_p50Latency, _p75Latency, _p90Latency, _p99Latency))
    file.write("Tot request:%d\tsucc:%d\tfail:%d\n" %
               (totalNum + failRequest, totalNum, failRequest))


def testExec(provider, appName, srcPath, memory, param, resFile):
    execLatencies = []
    # Prewarm
    cli_scripts.flush(appName, srcPath, provider, memorySize=memory)
    cli_scripts.invoke([appName] + param)
    failRequest = 0
    for i in range(TEST_TIME):
        time.sleep(1)
        if (i + 1) % max((TEST_TIME) // 10, 1) == 0:
            print(f"Test Step:{(i+1)}/{TEST_TIME}")
        try:
            rawResult = cli_scripts.invoke([appName] + param)
            startTime, returnTime = parseResult(rawResult)
            execLatency = returnTime - startTime
            execLatencies.append(execLatency)
        except:
            print(rawResult)
            failRequest += 1
    statResult(execLatencies, failRequest, "Execution", resFile)
    return


if __name__ == '__main__':
    params = sys.argv[1:]
    args = argparse.init_args(params)
    resultDir = args.resultDir
    f = open(resultDir + "/testcase1-Execution-Time.txt", 'a')
    srcPath = args.srcPath
    memory = args.memory
    param = args.param
    appName = args.appName
    testCaseDir = args.testCaseDir
    provider = args.provider
    resultDir = args.resultDir

    invokeParam = []
    for k, v in param.items():
        invokeParam.append(k)
        invokeParam.append(v)

    f.write("testcase1 for app %s. \n" % appName)
    testExec(provider, appName, srcPath, memory, invokeParam, f)
    f.write("---------------------------------------------\n")
    f.close()
