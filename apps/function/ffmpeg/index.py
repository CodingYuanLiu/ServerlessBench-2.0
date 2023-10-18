import os
import subprocess


def convert_video(inputFile, thumb_width, param):
    message = f'converting video successful. | src: {inputFile}. '
    try:
        command = [
            './ffmpeg', '-loglevel', 'error', '-y', '-i', inputFile, '-vf',
            f"scale={thumb_width}:-1", param["outputfile"]
        ]
        res = subprocess.run(command, check=True, stdout=subprocess.PIPE)
        message += res.stdout.decode()
    except Exception as e:
        message = f'converting video failed. | src: {inputFile} | Exception: {e}'
    return message


def handler(event, context={}):
    inputfile = event["payload"]
    thumb_width = event["thumb_width"]
    msg = convert_video(inputfile, thumb_width, event)
    return {"res": msg}
