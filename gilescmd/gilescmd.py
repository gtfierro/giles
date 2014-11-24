import argparse
import base64
import sys
import os
import glob
import time
import datetime
import json
import csv
import uuid
import pandas as pd
from smap.contrib import dtutil
from pymongo import MongoClient
from smap.contrib import dtutil
from smap.archiver.client import SmapClient
from collections import defaultdict

parser = argparse.ArgumentParser(description='Command line help tool for Giles')

subparsers = parser.add_subparsers(dest='subparsername', help='sub-command help')

parser_newkey = subparsers.add_parser('newkey', help='newkey help')
parser_newkey.add_argument('name',type=str,help='name help')
parser_newkey.add_argument('email',type=str,help='email help')
parser_newkey.add_argument('-p','--private',action='store_false',help='Makes this API key private')
parser_newkey.add_argument('-m', '--mongourl', default='localhost:27017', help='Mongo URL')

parser_streamid = subparsers.add_parser('streamid', help='streamid help')
parser_streamid.add_argument('apikey',type=str,help='apikey help')
parser_streamid.add_argument('uuid',type=str,help='uuid help')
parser_streamid.add_argument('-m', '--mongourl', default='localhost:27017', help='Mongo URL')

parser_rdbdump = subparsers.add_parser('rdbdump', help='Dump ReadingDB streams')
parser_rdbdump.add_argument('-r','--readingdb',default='localhost:4242',help='ReadingDB URL')
parser_rdbdump.add_argument('-s','--startyear',help='Year to start pulling data from',type=int,default=2010)
parser_rdbdump.add_argument('-e','--endyear',help='Year to stop pulling data from',type=int,default=2014)
parser_rdbdump.add_argument('-b','--blocksize',help='How many streamids to poll at once',type=int,default=10000)
parser_rdbdump.add_argument('-d','--directory',help='Directory to dump data',type=str,default='.')
parser_rdbdump.add_argument('-f','--streamfile',help='CSV file where each line is streamid,uuid. If specified, dumped files will be stored with UUID as filename. Else, streamid used as filename')

parser_smapdump = subparsers.add_parser('smapdump', help='Dump streams from sMAP archiver')
parser_smapdump.add_argument('-s','--smapurl', default='localhost:8079',help='sMAP Archiver URL')
parser_smapdump.add_argument('-q','--query', default='select distinct uuid',help='UUID filter query')

parser_import = subparsers.add_parser('import', help='Import data from external file into sMAP')
parser_import.add_argument('-d', '--delimiter', type=str, default=',', help='Expects each line in the file to be <timestamp><delim><value>. Allows you to specify a custom delimiter')
parser_import.add_argument('-e', '--header', action="store_true", help='Do the specified files include a header? If so, importer skips the first line')
parser_import.add_argument('-m', '--metadata', type=str, help='Metadata file containing k/v = uuid/metadata. Used to lookup Path names')
parser_import.add_argument('files', nargs='+', type=str, help='File(s) or directory to import. If directory, gilescmd will go 1 level deep and find all files. List of files can also be specified (use commas to separate files)')

args = parser.parse_args()
print args

# readingDB export convenience fxns
def get_times(startyear, endyear):
    for year in range(startyear,endyear+1):
        for month in range(1,13):
            start = "{0}-{1}-{2}".format(month, 1, year)
            if month < 12:
                end = "{0}-{1}-{2}".format(month+1, 1, year)
            else:
                end = "{0}-{1}-{2}".format(month, 31, year)
            start, end = map(lambda x: dtutil.dt2ts(dtutil.strptime_tz(x, '%m-%d-%Y')), (start, end))
            yield start, end-1

def get_streamids(blocksize):
    for i in range(101,500):
        print 'block {0} of {1}'.format(i - 101, 500 - 101)
        yield range(blocksize*i+1, blocksize*i+1+blocksize)

def get_streamids_file(filename, blocksize):
    with open(filename) as f:
        reader = csv.reader(f)
        counter = 0
        while reader:
            ids = []
            counter += 1
            for i in range(blocksize):
                ids.append(int(reader.next()[0]))
            print 'block',counter
            yield ids

if args.subparsername == 'newkey':
    url,port = args.mongourl.split(':')
    client = MongoClient(url,int(port))
    db = client.archiver
    name = args.name
    email = args.email
    randbytes = open('/dev/urandom').read(64)
    apikey = base64.urlsafe_b64encode(randbytes)
    db.apikeys.insert({'key': apikey, 'name': name, 'email': email, 'public': not args.private})
    print apikey
    sys.exit(0)

elif args.subparsername == 'streamid':
    url,port = args.mongourl.split(':')
    client = MongoClient(url,int(port))
    db = client.archiver
    apikey = args.apikey
    uuid = args.uuid
    res = db.metadata.find_one({'uuid': uuid, '_api': apikey})
    if not res:
        print 'No UUID {0} found for API key {1}'.format(uuid, apikey)
        sys.exit(1)
    res = db.streams.find_one({'uuid': uuid})
    print res['streamid']
    sys.exit(0)

elif args.subparsername == 'rdbdump':
    import readingdb as rdb
    url,port = args.readingdb.split(':')
    rdb.db_setup(url,int(port))
    db = rdb.db_open(host=url,port=int(port))
    streamfile = pd.read_csv(args.streamfile, header=None) if args.streamfile else None
    streamiter = get_streamids(args.blocksize) if not args.streamfile else get_streamids_file(args.streamfile, args.blocksize)
    for streamids in streamiter:
        for start,end in get_times(args.startyear,args.endyear):
            print "pulling {0} to {1} for streamids {2} to {3}".format(start, end, streamids[0], streamids[-1]),
            sys.stdout.flush()
            s = time.time()
            gotdata = False
            while not gotdata:
                try:
                    data = rdb.db_query(streamids, start, end)
                except Exception as e:
                    print 'error', e
                    continue
                gotdata = True
            e = time.time()
            print "in {:.3f} seconds".format(e - s),
            sys.stdout.flush()
            for sid, tsdata in zip(streamids, data):
                if len(tsdata) > 0:
                    d = pd.DataFrame(tsdata)
                    if args.streamfile:
                        sid = streamfile[streamfile[0] == sid][1].values[0]
                    with open('{0}/{1}.csv'.format(args.directory, sid), 'a+') as f:
                        d.to_csv(f, index=False, header=None)
            s = time.time()
            print "written to file in {:.3f} seconds".format(s - e)
    sys.exit(0)

elif args.subparsername == 'smapdump':
    client = SmapClient('http://'+args.smapurl)
    uuids = filter(lambda x: x, client.query(args.query))
    md = {}
    for uuid in uuids:
        md[uuid] = client.tags('uuid = "{0}"'.format(uuid))[0]

    begin = int(time.time())
    limit = float('inf')
    i = 0
    while limit > 0:
        data = []
        print datetime.datetime.now(), i
        end = begin- 60*60*24*i
        start = end - 60*60*24*30*(i+1)
        limit = start
        i += 1
        res = client.data_uuid(uuids, start, end)
        if not any(map(any, res)):
            print 'no more data!'
            break
        for uuid, uuiddata in zip(uuids, res):
            d = pd.DataFrame(uuiddata)
            d[0] = d[0] * 1000
            d[0] = d[0].astype(int)
            if not d[0].any(): continue
            with open(uuid+'.csv', 'a+') as f:
                d.to_csv(f, index=False, header=None)
    json.dump(md, open('metadata.json','w+'))

elif args.subparsername == 'import':
    import requests
    if args.files is None:
        print 'You must specify file(s) to import'
        sys.exit(1)
    files = args.files
    exists = map(lambda x: os.path.exists(x), files)
    if not all(exists):
        notexist = [y for x,y in zip(exists, files) if not x]
        print 'The following specified files do not exist: {0}'.format(','.join(notexist))
        sys.exit(1)
    for f in files:
        print "Reading",f
        if not args.header:
            d = pd.read_csv(f,sep=args.delimiter,header=None)
        else:
            d = pd.read_csv(f,sep=args.delimiter)
        obj = {'/sensor1': {'Readings': [], 'uuid': str(uuid.uuid1())}}
        row = 0
        while row*1000 < len(d):
            data = d[row*1000:(row+1)*1000].to_json(orient='values').replace('.0,',',')
            if not len(data):
                break
            obj['/sensor1']['Readings'] = json.loads(data)
            row += 1
            try:
                resp = requests.post('http://localhost:8079/add/jm-5tEwYdB39T-2cqYwM94kkRJ2-wQ0aNMSmflsjNidsuqBvlA4EtyMSTCYX5VEVhXIyvXFSlrB6dVIfoEIZVg==',data=json.dumps(obj))
                if not resp.ok:
                    print resp.content
                    print obj
            except:
                break

