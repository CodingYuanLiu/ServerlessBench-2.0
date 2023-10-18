import indexUserFunc
import time

def getMillis():
    return round(time.time() * 1000)

def handler (event, context={}):
   startTime = getMillis()
   res = indexUserFunc.handler(event, context)
   return  {'startTime': startTime}


