
def handler(event, context={}):
    q = len(event.get('text'))
    return {'number': q}
