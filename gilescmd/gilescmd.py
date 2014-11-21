import argparse
import base64
import sys
import time
import csv
from smap.contrib import dtutil
from pymongo import MongoClient
from smap.contrib import dtutil

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

parser_smapdump = subparsers.add_parser('smapdump', help='Dump streams from sMAP archiver')
parser_smapdump.add_argument('-s','--smapurl', default='localhost:8079',help='sMAP Archiver URL')
parser_smapdump.add_argument('-q','--query', default='select distinct uuid',help='UUID filter query')

args = parser.parse_args()
print args

# readingDB convenience fxns
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
    import pandas as pd
    import readingdb as rdb
    url,port = args.readingdb.split(':')
    rdb.db_setup(url,int(port))
    db = rdb.db_open(host=url,port=int(port))
    for streamids in get_streamids(args.blocksize): #get_streamids_file('streamids.csv',args.blocksize):
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
                    with open('{0}/{1}.csv'.format(args.directory, sid), 'a+') as f:
                        d.to_csv(f, index=False, header=None)
            s = time.time()
            print "written to file in {:.3f} seconds".format(s - e)
    sys.exit(0)

elif args.subparsername == 'smapdump':
    import pandas as pd
    from smap.archiver.client import SmapClient
    import time
    import datetime
    import json
    import pandas as pd
    from collections import defaultdict
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
            #d[0] = d[0].astype(str).replace('.0','')
            d[0] = d[0].astype(int)
            #d[0] = pd.to_datetime(d[0])
            if not d[0].any(): continue
            with open(uuid+'.csv', 'a+') as f:
                d.to_csv(f, index=False, header=None)
    json.dump(md, open('metadata.json','w+'))
