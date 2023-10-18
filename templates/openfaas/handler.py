from function import index
# args type : str
def handle(args):
    """Main."""
    if args != "":
        event = eval(args)
        rawResult = index.handler(event, {})
        result = ''
        if type(rawResult) == str:
            result =  {'result': rawResult}
        elif type(rawResult) == dict:
            result =  rawResult
        else:
            result = {'error': 'error return type'}
        return result
    return "args is null"

# rawResult = index.handler("args", {})
# result = ''
# if type(rawResult) == str:
#     result =  {'result': rawResult}
# elif type(rawResult) == dict:
#     result =  rawResult
# else:
#     result = {'error': 'error return type'}