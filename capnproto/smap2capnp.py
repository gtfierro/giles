import capnp
import json
import smap_capnp
import glob

jsondata = """
{
    "/": {
        "Contents": [
            "fast"
        ]
    },
    "/fast": {
        "Contents": [
            "sensor0"
        ]
    },
    "/fast/sensor0": {
        "Properties": {
            "ReadingType": "long",
            "Timezone": "America/Los_Angeles",
            "UnitofMeasure": "V",
            "UnitofTime": "s"
        },
        "Metadata": {
            "Site": "Test Site",
            "Nested": {
                "key": "value",
                "other": "value"
            }
        },
        "Readings": [[9182731928374, 30]],
        "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
    }
}
"""
jsonobj = json.loads(jsondata)

for path, contents in jsonobj.iteritems():
    print '#'*5,'New Message','#'*5
    print 'path',path
    print 'uuid', contents.get('uuid')
    print 'contents', contents.get('Contents')
    print 'properties', contents.get('Properties')
    print 'readings', contents.get('Readings')
    print 'metadata', contents.get('Metadata')
    msg = smap_capnp.Message.new_message()
    msg.path = path
    msg.uuid = bytes(contents.get('uuid'))
    if contents.get('Contents'):
        msg_contents = msg.init('contents', len(contents.get('Contents')))
        for i,item in enumerate(contents.get('Contents')):
            msg_contents[i] = item
    if contents.get('Readings'):
        msg_readings = msg.init('readings', len(contents.get('Readings')))
        for i, item in enumerate(contents.get('Readings')):
            msg_readings[i] = smap_capnp.Message.Reading.new_message(time= item[0], data= item[1])
    if contents.get('Properties'):
        msg_properties = msg.init('properties', len(contents.get('Properties')))
        for i, kv in enumerate(contents.get('Properties').iteritems()):
            msg_properties[i] = smap_capnp.Message.Pair.new_message(key = kv[0], value = kv[1])
    if contents.get('Metadata'):
        msg_metadata = msg.init('metadata', len(contents.get('Metadata')))
        for i, kv in enumerate(contents.get('Metadata').iteritems()):
            msg_metadata[i] = smap_capnp.Message.Pair.new_message(key = kv[0], value = kv[1])
    with open('{0}.bin'.format(path.replace('/','_')),'w+b') as f:
        msg.write(f)

print 'now reading back'

for filename in glob.glob('*.bin'):
    msg = smap_capnp.Message.read(open(filename,'rb'))
    print msg.to_dict()
