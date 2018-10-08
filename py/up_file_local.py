#!/usr/bin/env python

# http://flask.pocoo.org/docs/0.12/patterns/fileuploads/

# https://github.com/teamsoo/flask-api-upload-image/blob/master/server.py
# http://www.patricksoftwareblog.com/receiving-files-with-a-flask-rest-api/

from flask import Flask, url_for, send_from_directory, request, abort
import logging
import os
import docker
from werkzeug import secure_filename
from datetime import datetime
import json
import inotify.adapters

import Queue as queue
from threading import Event, Thread

# ./up_file_local.py && rm -rf uploadsPY/

# IDEA
# test: file x.txt - size OK, file y.txt - too large size
# it will create ./uploadsPY dir if not exists
# 'curl -X POST -F..' uploads each file to ./uploadsPY dir
# 'curl -X GET /getup/uploadsPY' displays ./uploadsPY dir content
# if 'go' or 'py' file uploaded it will run docker and display output on webpage
# .py should have SHEBANG to be validated successfully
# container should be stopped: curl .../stop/15ad

# REQUIREMENTS
# python file should have shebang and .py
# golang file (TOBEDONE)


"""
#TODO
- when app is running and uploadsPY will be removed -> error, it's because of this bug:
https://github.com/dsoprea/PyInotify/issues/51

- python cache

- watcher, ustawic czas sprawdzanie sciezki
- jak zainstalowac dodatkowe paczki jesli skrypt tego wymaga -> pyfiles i .whl
- wystawic HOST_PORT i DOCKER_PORT jako zmienna jesli wrzuce ten plik w kontener
- simple authentication
- zamiast s3 mozna korzystac z volume(workspace, cinder)
- nie udalo sie przeslac pliku, co zrobic?
"""

app = Flask(__name__)
SERVICE_PORT = 5000
HOST_PORT = 8000
DOCKER_PORT = 1111
DOCKER_ID = ''
GETNEWFILE = ''

file_handler = logging.FileHandler('/tmp/server.log')
app.logger.addHandler(file_handler)
app.logger.setLevel(logging.DEBUG)

PROJECT_HOME = os.path.dirname(os.path.realpath(__file__))
UPLOAD_FOLDER = 'uploadsPY'
UPLOAD_CATALOG = '{}/{}/'.format(PROJECT_HOME, UPLOAD_FOLDER)

ALLOWED_EXTENSIONS = set(['txt', 'py', 'go'])

# http://flask.pocoo.org/docs/1.0/config/
# prefix to all flask routes, if not setup default is '/'
app.config['APPLICATION_ROOT'] = '/api/v1'
app.config['UPLOAD_CATALOG'] = UPLOAD_CATALOG
# maximum allowed payload to 5 megabytes (file size)
app.config['MAX_CONTENT_LENGTH'] = 5 * 1024 * 1024

# run docker container if new file
def runDocker(fileName, fileType):
	# import docker
	# client = docker.from_env()
	if fileType == 'py':
		dcid = os.popen("docker run -d -v {0}/{1}:/app/{1} \
		--env APPNAME={1} \--env PORT={2} -p {3}:{2}  \
		local/py_faas:0.0.1".format(UPLOAD_CATALOG, fileName, DOCKER_PORT, HOST_PORT))
		global DOCKER_ID
		DOCKER_ID = dcid.read()
		# DOCKER_ID = "123"
		return DOCKER_ID
	elif fileType == 'go':
		print('gogo to be done')

# monitor UPLOAD_CATALOG directory
def fire():
	# notifier = inotify.adapters.Inotify()
	# notifier.add_watch(UPLOAD_CATALOG)
	notifier = inotify.adapters.InotifyTree(PROJECT_HOME)
	for event in notifier.event_gen():
		if event is not None:
			# print(event)      # uncomment to see all events generated
			if 'IN_CREATE' in event[1]:
				print("[+] File '{0}' created in '{1}'".format(event[3], event[2]))
				global GETNEWFILE
				GETNEWFILE = event[3]

def create_new_dir(local_dir):
    newpath = local_dir
    if not os.path.exists(newpath):
        os.makedirs(newpath)
    return newpath

# get a list of uploaded files
# @app.route('/api/v1/getup/<folder>', methods = ['GET'])
@app.route('/getup/<folder>', methods = ['GET'])
def apiGetup(folder = None):
	# curl -v localhost:5000/getup/uploadsPY
	app.logger.info(PROJECT_HOME)
	if folder is None or folder != UPLOAD_FOLDER:
		abort(404)
	else:
		var = os.listdir(UPLOAD_CATALOG)
		return json.dumps(var)

# upload file and determine file type
# @app.route('/api/v1/up', methods = ['POST'])
@app.route('/up', methods = ['POST'])
def apiUpload():
	# curl -X POST -F 'file=@x.txt' localhost:5000/up

	app.logger.info(PROJECT_HOME)
	TIME_ST = datetime.now().strftime("%y%m%d%H%M%S")
	# gdzie 'file' to pole, moze byc 'image' itp
	if request.method == 'POST' and request.files['file']:
		app.logger.info(app.config['UPLOAD_CATALOG'])

		files = request.files['file']
		fileName = secure_filename(files.filename)

		# fire() wymaga zeby ta sciezka istniala wczesniej
		# create_new_dir(app.config['UPLOAD_CATALOG'])
		
		saved_path = os.path.join(app.config['UPLOAD_CATALOG'], \
									"{}-".format(TIME_ST) + fileName)
		
		app.logger.info("saving {}".format(saved_path))
		
		files.save(saved_path)

		# display file content
		# http://flask.pocoo.org/docs/1.0/api/#flask.send_file
		# as_attachment=True?
		# return send_from_directory(app.config['UPLOAD_CATALOG'], \
		# 							"{}-".format(TIME_ST) + fileName)
		msg = "[+] File uploaded successfully: {}\n".format(fileName)

		# DETERMINE FILE TYPE
		name, ext = os.path.splitext(GETNEWFILE)
		if ext == '.py' and \
		'python' in os.popen(
			"file {}/{}".format(UPLOAD_CATALOG, GETNEWFILE)).read().lower():
			runDocker(GETNEWFILE, 'py')
			# print("runDocker py..")
			return msg + "[+] Docker ID: {}\n".format(DOCKER_ID)
		# elif ext == '.go' and 'c source' in \
		# os.popen("file {}/{}".format(UPLOAD_CATALOG, event[3])).read().lower():
		# 	runDocker(event[3], 'go')			
		return msg

	else:
		return "Hey file, where are you?"

# stop running container
# @app.route('/api/v1/stop/<id>', methods = ['GET'])
@app.route('/stop/<id>', methods = ['GET'])
def apiStop(id = None):
	app.logger.info(PROJECT_HOME)
	# curl localhost:5000/stop/07c28
	if id is None:
		abort(404)
	else:
		# var = os.popen("docker stop {}".format(id)).read()
		# return "Docker container stopped, id " + var
		client = docker.from_env()
		container = client.containers.get("{}".format(id))
		container.stop()
		return "Docker container stopped, id: {}\n".format(id)

def main():
	# fire() wymaga zeby ta sciezka istniala wczesniej niz przy apiUpload()
	# create_new_dir() zakomentowane
	if not os.path.exists(UPLOAD_CATALOG):
		os.makedirs(UPLOAD_CATALOG)
	q = queue.Queue()
	threads = Thread(target=fire)
	threads.daemon = True
	threads.start()
	app.run(host = '0.0.0.0', port = SERVICE_PORT, debug = False)

if __name__ == '__main__':
	main()
	# try:
	# 	...
	# except KeyboardInterrupt:
	# 	print('\nCtrl+C - Stopping server')
	# 	sys.exit(1)
		
