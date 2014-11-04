import argparse
import base64
import sys
from pymongo import MongoClient

parser = argparse.ArgumentParser(description='Command line help tool for Giles')
parser.add_argument('-m', '--mongourl', default='localhost:27017', help='Mongo URL')

subparsers = parser.add_subparsers(dest='subparsername', help='sub-command help')

parser_newkey = subparsers.add_parser('newkey', help='newkey help')
parser_newkey.add_argument('name',type=str,help='name help')
parser_newkey.add_argument('email',type=str,help='email help')
parser_newkey.add_argument('-p','--private',action='store_false',help='Makes this API key private')

parser_streamid = subparsers.add_parser('streamid', help='streamid help')
parser_streamid.add_argument('apikey',type=str,help='apikey help')
parser_streamid.add_argument('uuid',type=str,help='uuid help')

args = parser.parse_args()
print args

url,port = args.mongourl.split(':')
client = MongoClient(url,int(port))
db = client.archiver
#col = db.collectionname
#col.insert(doc)
#col.find(doc), find_one(doc)

if args.subparsername == 'newkey':
    name = args.name
    email = args.email
    randbytes = open('/dev/urandom').read(64)
    apikey = base64.urlsafe_b64encode(randbytes)
    db.apikeys.insert({'key': apikey, 'name': name, 'email': email, 'public': not args.private})
    print apikey
    sys.exit(0)

elif args.subparsername == 'streamid':
    apikey = args.apikey
    uuid = args.uuid
    res = db.metadata.find_one({'uuid': uuid, '_api': apikey})
    if not res:
        print 'No UUID {0} found for API key {1}'.format(uuid, apikey)
        sys.exit(1)
    res = db.streams.find_one({'uuid': uuid})
    print res['streamid']
    sys.exit(0)
