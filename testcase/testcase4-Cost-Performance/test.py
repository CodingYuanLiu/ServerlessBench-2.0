import os
import sys
import time
import datetime
import subprocess

os.path.dirname(__file__)
sys.path.append(os.path.join(os.path.dirname(__file__), ".."))
from util import argparse
from util import cli_scripts

TEST_TIME = 1  # traverse all memory size
EPOCH_TEST_TIME = 5  #invoke times
MEM_RANGE = 1  #max tested mem size(x * 128M), max:21
WARM_TIME = 10
WAIT_TIME = 0.5


def parseResult(rawResult):
    startTime = int(eval(rawResult[0])['startTime'])
    returnTime = int(eval(rawResult[0])['returnTime'])
    invokeTime = int(rawResult[1])
    endTime = int(rawResult[2])
    return invokeTime, startTime, returnTime, endTime


def statResult(allLatencies, failRequest, memorySize, file):
    content = "\t%dM | " % memorySize
    for i in range(3):
        latencies = allLatencies[i]
        latencies.sort()
        totalNum = len(latencies)
        if totalNum == 0:
            if i == 0:
                cost = 0
            content += "all request fail |,"
            continue
        _p50Latency = latencies[int(totalNum * 0.5) - 1]
        _p75Latency = latencies[int(totalNum * 0.75) - 1]
        _p90Latency = latencies[int(totalNum * 0.9) - 1]
        _p99Latency = latencies[int(totalNum * 0.99) - 1]
        if i == 0:
            cost = _p90Latency * memorySize / 1024
        content += "%d, %d, %d, %d, |, " % (_p50Latency, _p75Latency,
                                            _p90Latency, _p99Latency)
    content += str(cost) + "\n"
    file.write(content)
    file.write("Tot request:%d\tsucc:%d\tfail:%d\n" %
               (totalNum + failRequest, totalNum, failRequest))


def testPrice(appName, param, resFile, memorySize):
    execLatencies = []
    startLatencies = []
    e2eLatencies = []

    # Prewarm
    cli_scripts.preWarm(appName, param, WARM_TIME, WAIT_TIME)

    i = 0
    failRequest = 0
    while i < EPOCH_TEST_TIME:
        i += 1
        try:
            rawResult = cli_scripts.invoke([appName] + param)
            invokeTime, startTime, returnTime, endTime = parseResult(rawResult)
        except:
            failRequest += 1
            continue
        execLatency = returnTime - startTime
        startLatency = startTime - invokeTime
        e2eLatency = endTime - invokeTime
        execLatencies.append(execLatency)
        startLatencies.append(startLatency)
        e2eLatencies.append(e2eLatency)
    allLatancies = []
    allLatancies.append(execLatencies)
    allLatancies.append(startLatencies)
    allLatancies.append(e2eLatencies)

    statResult(allLatancies, failRequest, memorySize, resFile)
    return


if __name__ == '__main__':
    params = sys.argv[1:]
    args = argparse.init_args(params)
    srcPath = args.srcPath
    memory = args.memory
    param = args.param
    appName = args.appName
    testCaseDir = args.testCaseDir
    provider = args.provider
    resultDir = args.resultDir
    f = open(resultDir + "/testcase4-Cost-Performance.txt", 'a')
    invokeParam = []
    for k, v in param.items():
        invokeParam.append(k)
        invokeParam.append(v)

    f.write("testcase4 for app %s. \n" % appName)
    f.write("\tCostPrice result:\n")
    f.write(
        "\tmemory size | exec latency | start latency | e2e latency | cost\n")
    for i in range(TEST_TIME):
        if (i + 1) % max((TEST_TIME) // 10, 1) == 0:
            print(f"Test Step:{(i+1)}/{TEST_TIME}")
        memorySizeIndexRange = range(1, min(21, MEM_RANGE + 1), 1)
        for memorySizeIndex in memorySizeIndexRange:
            memorySize = memorySizeIndex * 512
            print(f"MemSize:{memorySize}")
            # modify the ./flush.sh file to set memorySize
            cli_scripts.flush(appName,
                              srcPath,
                              provider,
                              memorySize=str(memorySize))
            testPrice(appName, invokeParam, f, memorySize)
    f.write("---------------------------------------------\n")
    f.close()
