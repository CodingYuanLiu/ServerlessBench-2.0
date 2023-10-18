import importlib
import json
class bucket:
    client_b = None
    p = ""
    bucketname = ""

    def __init__(self, context, provider, bucket_name):
        self.p = provider
        self.bucketname = bucket_name
        if provider == 'ali':
            f = open('.serverless.config','r')
        elif provider == 'huawei':
            f = open('code/.serverless.config','r')
        else:
            f = None
        data = json.load(f)
        AK = data['AK']
        SK = data['SK']

        if( provider == "huawei"):
            obs = importlib.import_module("obs")
            server='https://obs.cn-east-3.myhuaweicloud.com'
            self.client_b = obs.ObsClient(access_key_id=AK, secret_access_key=SK, server=server,
            path_style=True, region='china', ssl_verify=False, max_retry_count=5, timeout=20)         
            #self.client_b = obsClient.bucketClient(bucket_name)
        elif(provider == "ali"):
            oss2 = importlib.import_module("oss2")
            endpoint = 'http://oss-cn-hangzhou.aliyuncs.com' 
            auth = oss2.Auth(AK, SK)
            self.client_b = oss2.Bucket(auth, endpoint, bucket_name)
        else:
            pass

    def Get(self, key):
        if self.p == 'ali':
            return self.client_b.get_object(key) 
        elif self.p == 'huawei':
            return self.client_b.getObject(self.bucketname, key, loadStreamInMemory=True)
        else:
            return None
    def Set(self, key, value):
        if self.p == 'ali':
            return self.client_b.put_object(key, value)
        elif self.p == 'huawei':
            return self.client_b.putContent( self.bucketname, key, value)
        else:
            return None

        
