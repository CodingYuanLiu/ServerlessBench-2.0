def handler(event, context={}):
    return {"text": event.get('text') + ":hello world!"}
