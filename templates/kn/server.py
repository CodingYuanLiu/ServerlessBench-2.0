from flask import Flask
from flask import request

import index
import logging
import json
import os 
app = Flask(__name__)

@app.route("/", methods=['GET'])
def hello_world():
    event = {}
    if request.data:
        event = request.get_json()
    context = {}
    rawResult = index.handler(event, context)
    result = ''
    if type(rawResult) == str:
        result =  {'result': rawResult}
    elif type(rawResult) == dict:
        result =  rawResult
    else:
        result = {'error': 'error return type'}
    return result

@app.route("/", methods=['POST'])
def pubsub_push():
    print(request.data)
    event = json.loads(request.data.decode('utf-8'))
    context = {}
    rawResult = index.handler(event, context)
    result = ''
    if type(rawResult) == str:
        result =  {'result': rawResult}
    elif type(rawResult) == dict:
        result =  rawResult
    else:
        result = {'error': 'error return type'}
    app.logger.info(f'Workflow return:\n{result}')
    return 'OK', 200


if __name__ != '__main__':
    # Redirect Flask logs to Gunicorn logs
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)
else:
    app.run(debug=True, host='0.0.0.0', port=int(os.environ.get('PORT', 8080)))
