#!/usr/bin/env python

# powiazane ze 'spark_api_s3.py'

try:
    import boto
    import boto.s3.connection
    from boto.s3.key import Key
except ImportError:
    raise ImportError("pip 'boto' package is missing")

ACCESS_KEY = '<access_key>'
SECRET_KEY = '<secret_key>'
ENDPOINT_URL = '<endpoint_url>'
ENDPOINT_PORT = '<port>'
BUCKET_NAME = '<bucket_name>'

INPUT_FILE_NAME = '<filename>'
OUTPUT_FILE_NAME = '<filename>'
DOWNLOAD_TO = '<full-path-directory>' # Ex: '/data/folder/'
UPLOAD_FROM = '<full-path-directory>' # Ex: '/data/folder2/'

#OPENSTACK
conn = boto.connect_s3 (
                        aws_access_key_id = ACCESS_KEY, 
                        aws_secret_access_key = SECRET_KEY, 
                        host = ENDPOINT_URL,
                        port = int(ENDPOINT_PORT), 
                        is_secure=False,     # comment if you are using ssl
                        calling_format = boto.s3.connection.OrdinaryCallingFormat()
                        )

PROJECT_S3 = 's3://{}/'.format(BUCKET_NAME)
UPLOAD_FOLDER = '{}/uploads/'.format(PROJECT_HOME)
ALLOWED_EXTENSIONS = set(['txt', 'pdf', 'png', 'jpg', 'jpeg', 'gif'])

app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER

# check if bucket exists?
# check if upload_folder exists if not create
