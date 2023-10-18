import requests
import json


def handler(event, context={}):
    url = event.get("url", "http://httpbin.org/post")
    headers = {'Accept': 'application/json'}
    payload = json.loads(event["payload"])

    r = requests.post(url, headers=headers, data=payload)

    msg = str({"status": r.status_code, "content": r.content})

    return {"res": msg}
