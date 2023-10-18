import indexUserFunc
import time
import json

def getMillis():
    return round(time.time() * 1000)

def handler (event, context={}):
    savedParam = eval(event.get("savedParam", "[]"))

    startTime = getMillis()
    res = indexUserFunc.handler(event, context)
    returnTime = getMillis()

    stageParamList = []
    for key, value in res.items():
        stageParamList += [str(key), str(value)]
    savedParam += [stageParamList]

    timeStampList = eval(event.get("timeStamp", "[]"))
    timeStampList += [startTime, returnTime]
    res["timeStamp"] = str(timeStampList)
    res["savedParam"] = str(savedParam)
    return json.dumps(res)
