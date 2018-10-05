#!/usr/bin/env python

from flask import Flask
import subprocess
import os
import json

app = Flask(__name__)

# SERVICE_PORT = os.getenv('PORT', 8000)
# APP_NAME = os.getenv('APP_NAME', 'z.py')

SERVICE_PORT = os.getenv('PORT')
APP_NAME = os.getenv('APPNAME', 'default.py')

@app.route("/")
def hello():
    cmd = ['python', APP_NAME]
    p = subprocess.Popen(cmd, 
                        stdout = subprocess.PIPE,
                        stderr=subprocess.PIPE,
                        stdin=subprocess.PIPE)
    out,err = p.communicate()

    if out == '':
        return err
    return out

def main():
	app.run(host = '0.0.0.0', port = SERVICE_PORT, debug = False)

if __name__ == '__main__':
    main()
