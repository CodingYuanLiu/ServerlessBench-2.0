import indexUserFunc
import time
import json

def getMillis():
    return round(time.time() * 1000)


def handler (event, context={}):
    startTime = getMillis()
    res = indexUserFunc.handler(event, context)
    returnTime = getMillis()
    timeStamp = event.get("timeStamp", "")
    if not timeStamp:
        timeStamp = "[]"
    timeStampList = eval(timeStamp)
    timeStampList += [startTime, returnTime]
    res["timeStamp"] = str(timeStampList)
    return json.dumps(res)

