import index
def main(args):
    """Main."""
    rawResult = index.handler(args, {})
    result = ''
    if type(rawResult) == str:
        result =  {'result': rawResult}
    elif type(rawResult) == dict:
        result =  rawResult
    else:
        result = {'error': 'error return type'}
    return result

