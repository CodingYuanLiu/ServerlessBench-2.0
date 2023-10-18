
import os
import subprocess
import time
import datetime
import signal

def getMillis():
    return round(time.time() * 1000)

def getPodInfo(provider,functionName,namespace = "default"):
  funcName = functionName
  if provider == "knative":
    funcName = functionName + "-bench"
  if provider == "openwhisk":
     funcName = "-" + functionName
  return os.popen("kubectl get pod -n %s 2>/dev/null | grep -- %s"%(namespace,funcName)).read().split('\n')

def checkFunc(provider,functionName, needCold=False, namespace = "default"):
  runningFlag=False
  scaleToZeroFlag=False
  while not((needCold and scaleToZeroFlag==True) or (needCold==False and runningFlag==True)):
    time.sleep(1)
    scaleToZeroFlag=True
    runningFlag=True
    for podInfo in getPodInfo(provider,functionName,namespace):
      if len(podInfo)<=1:
        continue
      scaleToZeroFlag = False
      podName = podInfo.split()[0]
      podStatus = podInfo.split()[2]
      if podStatus!="Running":
        runningFlag=False
        break
  return True

def flush(appName, srcPath='', provider='', memorySize=128, needCold=False, isFlow=False, srcPathList='',reqPathList='', memList='', stageNameList='', show_outout=False):
   output = subprocess.PIPE if show_outout else subprocess.DEVNULL
   sleepSecond = "70" if needCold else "7"
   if isFlow:
       delete_cmd = "./cli delete -f "+appName
       create_cmd = "./cli create -f "+ appName + " -c " + srcPathList+ " -y "+ memList + " -q " + reqPathList + " -n " + stageNameList
   else:
       delete_cmd = "./cli delete "+ appName
       create_cmd = "./cli create -d "+ appName + " " + srcPath+ " --memory "+ str(memorySize)

   if provider == "knative":
      subprocess.run(delete_cmd, shell=True,stdout=output, stderr=output)
      subprocess.run("sleep 3", shell=True)
      subprocess.run("docker volume rm $(docker volume ls -qf dangling=true)", shell=True,stdout=output, stderr=output)
      subprocess.run("kubectl delete pod `kubectl get pods | grep Terminating | awk {'print $1'}` --grace-period=0 --force", shell=True,stdout=output, stderr=output)
      subprocess.run(create_cmd, shell=True,stdout=output,stderr=output)   
      time.sleep(7)
      checkFunc(provider,appName)
      if needCold:
        checkFunc(provider,appName,needCold=True)
      time.sleep(1)
   elif provider == "openfaas":
      subprocess.run(delete_cmd, shell=True)
      subprocess.run("docker volume rm $(docker volume ls -qf dangling=true)", shell=True)
      subprocess.run("sleep 10", shell=True)
      subprocess.run(create_cmd, shell=True)
   elif provider == "openwhisk":
      subprocess.run(delete_cmd, shell=True)
      subprocess.run(create_cmd, shell=True)
      time.sleep(7)
      checkFunc(provider,appName,namespace="openwhisk")
      if needCold:
        checkFunc(provider,appName,needCold=True,namespace="openwhisk")
      time.sleep(1)
   elif provider == "fission":
      subprocess.run(delete_cmd, shell=True)
      subprocess.run("sleep 3", shell=True)
      subprocess.run("kubectl delete pod `kubectl get pods -n fission-function  | grep Terminating | awk {'print $1'}` -n fission-function --grace-period=0 --force", shell=True)
      subprocess.run("docker volume rm $(docker volume ls -qf dangling=true)", shell=True)
      subprocess.run(create_cmd, shell=True)
      subprocess.run("sleep "+ sleepSecond, shell=True)
      pass
   else:
      raise KeyError("Invalid provider%s."%(provider))

def invoke(params, isFlow=False):
      if isFlow:
         invoke_cmd = ["./cli", "invoke", '-f'] + params
      else:
         invoke_cmd = ["./cli", "invoke"] + params
      invokeTime = int(round(time.time() * 1000))
      res = subprocess.run(invoke_cmd,check=True,stdout=subprocess.PIPE)
      endTime = int(round(time.time() * 1000))
      res = res.stdout.decode()
      return [res, invokeTime, endTime]

def preWarm(app_name, param, warm_time, wait_time, isFlow=False, show_outout=False):
   output= subprocess.PIPE if show_outout else subprocess.DEVNULL
   if isFlow:
      invoke_cmd = ["./cli", "invoke", '-f'] + [app_name] + param
   else:
      invoke_cmd = ["./cli", "invoke"] + [app_name] + param
   total_startTime = getMillis()
   while(getMillis() - total_startTime < warm_time * 1000):
      start = datetime.datetime.now()
      process = subprocess.Popen(invoke_cmd, stdout=output, stderr=output)
      while process.poll() is None:
            time.sleep(0.2)
            now = datetime.datetime.now()
            if (now - start).seconds > warm_time / 10:
                try:
                    os.kill(process.pid, signal.SIGKILL)
                    os.waitpid(-1, os.WNOHANG)
                except:
                    break
      time.sleep(wait_time)
