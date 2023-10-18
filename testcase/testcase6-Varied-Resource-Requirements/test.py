import os
import sys
import json
import re

os.path.dirname(__file__)
sys.path.append(os.path.join(os.path.dirname(__file__), ".."))
from util import argparse
from util import cli_scripts

TEST_TIME = 1  #traverse all memory size
EPOCH_TEST_TIME = 30  #invoke times
FAIL_TIME = 10
WARM_TIME = 10
WAIT_TIME = 0.5
MEM_RANGE = 1  #max memory range, x * 128 M.


def parseResult(rawResult, is_Flow):
    savedParam = []
    if is_Flow:
        resultDict = eval(rawResult[0])
        timeStamps = json.loads(resultDict["timeStamp"])
        savedParam = json.loads(resultDict["savedParam"])
    else:
        resultStr = rawResult[0].strip().replace('\n', '')
        resultDict = json.loads(eval(resultStr)['result'])
        timeStamps = eval(resultDict["timeStamp"])
        savedParam = eval(resultDict["savedParam"])
    return timeStamps[0], timeStamps[-1], savedParam


def statResult(latencies, memorySize, file):
    content = "\t\t%dM | " % memorySize
    latencies.sort()
    totalNum = len(latencies)
    _p50Latency = latencies[int(totalNum * 0.5) - 1]
    _p75Latency = latencies[int(totalNum * 0.75) - 1]
    _p90Latency = latencies[int(totalNum * 0.9) - 1]
    _p99Latency = latencies[int(totalNum * 0.99) - 1]
    cost = _p90Latency * memorySize / 1024
    content += "%d, %d, %d, %d, |, " % (_p50Latency, _p75Latency, _p90Latency,
                                        _p99Latency)
    content += str(cost) + "\n"
    file.write(content)


def testPrice(appName, param, resFile, memorySize, isFlow=False):
    execLatencies = []
    # Prewarm
    print("\tpreWarm.")
    cli_scripts.preWarm(appName, param, WARM_TIME, WAIT_TIME, isFlow=isFlow)
    i = 0
    fail_times = 0
    while i < EPOCH_TEST_TIME:
        if (i + 1) % max((EPOCH_TEST_TIME) // 10, 1) == 0:
            print(f"\tTest Step:{(i+1)}/{EPOCH_TEST_TIME}")
        if fail_times >= FAIL_TIME:
            break
        i += 1
        rawResult = cli_scripts.invoke([appName] + param, isFlow=isFlow)
        try:
            startTime, returnTime, savedParam = parseResult(rawResult, isFlow)
        except:
            print(
                f"Test failed. fail_times:{fail_times}/{FAIL_TIME}. Raw result:{rawResult}"
            )
            i -= 1
            fail_times += 1
            continue
        execLatency = returnTime - startTime
        execLatencies.append(execLatency)

    statResult(execLatencies, memorySize, resFile)
    return savedParam


if __name__ == '__main__':
    params = sys.argv[1:]
    args = argparse.init_args(params)
    srcPath = args.srcPathList
    reqPath = args.reqPathList
    memList = args.memSizeList
    nameList = args.stageNameList
    param = args.param
    appName = args.appName
    testCaseDir = args.testCaseDir
    provider = args.provider
    resultDir = args.resultDir
    f = open(resultDir + "/testcase6-Varied-Resource-Requirements.txt", 'a')
    firstInvokeParam = []
    for k, v in param.items():
        firstInvokeParam.append(k)
        firstInvokeParam.append(v)

    stageNames = nameList.split(',')
    stagePaths = srcPath.split(',')

    f.write("testcase6 for app %s. \n" % appName)
    f.write("\tWorkflow %s result:\n" % (appName))
    f.write("\t\tmemory size | exec latency | cost\n")
    print("Testing the whole workflow.")
    for i in range(TEST_TIME):
        print(f"Step:{i}/{TEST_TIME}")
        memorySizeIndexRange = range(1, MEM_RANGE + 1, 1)
        for memorySizeIndex in memorySizeIndexRange:
            memorySize = memorySizeIndex * 128
            memList = [str(memorySize) for _ in range(len(stageNames))]
            # modify the ./flush.sh file to set memorySize
            print("\tflush workflow.")
            cli_scripts.flush(appName,
                              provider=provider,
                              isFlow=True,
                              srcPathList=srcPath,
                              reqPathList=reqPath,
                              memList=','.join(memList),
                              stageNameList=nameList,
                              show_outout=True)
            savedParam = testPrice(appName,
                                   firstInvokeParam,
                                   f,
                                   memorySize,
                                   isFlow=True)
    print("Testing splited functions.")
    for stage in range(len(stageNames)):
        print(f"Function: {stageNames[stage]}.")
        f.write("\tSplited function %s result:\n" % (stageNames[stage]))
        f.write("\t\tmemory size | exec latency | cost\n")
        invokeParam = savedParam[stage - 1] if stage != 0 else firstInvokeParam
        for i in range(TEST_TIME):
            print(f"Step:{i}/{TEST_TIME}")
            memorySizeIndexRange = range(1, MEM_RANGE + 1, 1)
            for memorySizeIndex in memorySizeIndexRange:
                memorySize = memorySizeIndex * 128
                # modify the ./flush.sh file to set memorySize
                print("\tflush function.")
                cli_scripts.flush(stageNames[stage],
                                  stagePaths[stage],
                                  provider,
                                  memorySize=str(memorySize))
                testPrice(stageNames[stage], invokeParam, f, memorySize)
    f.write("---------------------------------------------\n")
    f.close()
