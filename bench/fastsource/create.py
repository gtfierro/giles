from ConfigParser import RawConfigParser
from StringIO import StringIO
import sys
import uuid

configfile="""
[report 0]
ReportDeliveryLocation = http://localhost:8079/add

[/]
uuid = db43b080-176c-11e4-b2ab-6003089ed1d0

[server]
port = 8080

[/fast]
type = fast.Fast
rate = .01
number = 2
"""

s = StringIO(configfile)

number = int(sys.argv[1])

source = RawConfigParser()
source.optionxform=str
source.readfp(s)

for i in xrange(number):
    print source.get('server','port')
    port = int(source.get('server','port'))
    source.set('server', 'port', port+1)
    source.set('/','uuid', str(uuid.uuid1()))
    source.write(open('fast{0}.ini'.format(i), 'wb'))

