import indexUserFunc
import time

def getMillis():
    return round(time.time() * 1000)

def handler (event, context={}):
   startTime = getMillis()
   indexUserFunc.handler(event, context)
   returnTime = getMillis()
   return  {'startTime': startTime, 'returnTime': returnTime}

