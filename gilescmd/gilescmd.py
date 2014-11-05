import argparse
import base64
import sys
from pymongo import MongoClient

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

parser_smapdump = subparsers.add_parser('smapdump', help='Dump streams from sMAP archiver')
parser_smapdump.add_argument('-s','--smapurl', default='localhost:8079',help='sMAP Archiver URL')
parser_smapdump.add_argument('-q','--query', default='select distinct uuid',help='UUID filter query')

args = parser.parse_args()
print args


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
    db = rdb.db_open(url)
    print 'not implemented yet'
    sys.exit(1)

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
