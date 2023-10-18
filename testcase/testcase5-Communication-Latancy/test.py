import os
import sys
import json
import statistics

os.path.dirname(__file__)
sys.path.append(os.path.join(os.path.dirname(__file__), ".."))
from util import argparse
from util import cli_scripts

TEST_TIME = 10
FAIL_TIME = 10
WARM_TIME = 10
WAIT_TIME = 0.5


def parseResult(rawResult):
    timeStamps = json.loads(eval(rawResult[0])["timeStamp"])
    communicationTimes = []
    for stage in range(1, int(len(timeStamps) / 2)):
        after = timeStamps[2 * stage]
        before = timeStamps[2 * stage - 1]
        communicationTimes.append(int(after) - int(before))
    return statistics.mean(communicationTimes)


def statResult(latencies, file):
    latencies.sort()
    totalNum = len(latencies)
    _p50Latency = latencies[int(totalNum * 0.5) - 1]
    _p75Latency = latencies[int(totalNum * 0.75) - 1]
    _p90Latency = latencies[int(totalNum * 0.9) - 1]
    _p99Latency = latencies[int(totalNum * 0.99) - 1]
    file.write("average communication latency (ms):\n")
    file.write("50%\t75%\t90%\t99%\n")
    file.write("%d\t%d\t%d\t%d\n" %
               (_p50Latency, _p75Latency, _p90Latency, _p99Latency))


def testCommunication(provider, appName, srcPath, reqPath, memList, nameList,
                      invokeParam, f):
    communicationLatencies = []
    print("flush.")
    cli_scripts.flush(appName,
                      provider=provider,
                      isFlow=True,
                      srcPathList=srcPath,
                      reqPathList=reqPath,
                      memList=memList,
                      stageNameList=nameList,
                      show_outout=True)
    print("preWarm.")
    cli_scripts.preWarm(appName,
                        invokeParam,
                        WARM_TIME,
                        WAIT_TIME,
                        isFlow=True)

    i = 0
    fail_times = 0
    while i < TEST_TIME:
        if (i + 1) % max((TEST_TIME) // 10, 1) == 0:
            print(f"Test Step:{(i+1)}/{TEST_TIME}")
        if fail_times >= FAIL_TIME:
            break
        i += 1
        rawResult = cli_scripts.invoke([appName] + invokeParam, isFlow=True)
        try:
            communicationTime = parseResult(rawResult)
        except:
            print(
                f"Test failed. fail_times:{fail_times}/{FAIL_TIME}. Raw result:{rawResult}"
            )
            i -= 1
            fail_times += 1
            continue

        communicationLatencies.append(communicationTime)
    statResult(communicationLatencies, f)
    return


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
    f = open(resultDir + "/testcase5-Communication-Latancy.txt", 'a')
    invokeParam = []
    for k, v in param.items():
        invokeParam.append(k)
        invokeParam.append(v)

    f.write("testcase5 for app %s. \n" % appName)
    testCommunication(provider, appName, srcPath, reqPath, memList, nameList,
                      invokeParam, f)
    f.write("---------------------------------------------\n")
    f.close()
