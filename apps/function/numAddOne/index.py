def getPrimes(max):
    return max + 1


def handler(event, context={}):
    q = event.get('number')
    return {'number': getPrimes(int(q))}
