from flask import request
import index

# 只能返回string，否则报错
def main():
    # 使用这种方式接收http param（string）
    args = request.get_data()
    # return str(args)
    event = {}
    if args != "":
        event = eval(args)
    # return "{\'event\': " + str(event) + "}"
    rawResult = index.handler(event, {})
    result = ""
    if type(rawResult) == str:
        result =  "{\'result\': " + rawResult + "}"
    elif type(rawResult) == dict:
        result =  str(rawResult)
    else:
        result = "{\'error\': \'error return type\'}"
    return result
    
