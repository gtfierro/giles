import capnp
import json
import smap_capnp
import glob
from smap.util import buildkv, build_recursive

"""
Need a python library to translate sMAP messages back and forth between JSON and Capnproto
"""

apikey = "30tznFNq7R-mPqFITaGT-nh3kt9v6bH7oY4gwioJQAm7wa1ik43oOPZDpA2CIAxoVK4Qn-ZO1F5ZTlLpXpgAsQ=="

def json2capnp(jsonobj):
    """
    Expecting a sMAP report json object where the toplevel keys are paths
    and the toplevel values are the usual sMAP objects, e.g.

    {
        "/fast": {
            "Contents": [
                "sensor0"
            ]
        },
        "/fast/sensor0": {
            "Properties": {
                "ReadingType": "long",
                ...
            },
            "Metadata": {
                "Site": "Test Site",
                ...
            },
            "Readings": [[9182731928374, 30]],
            "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
        }
    }

    """
    messages = []
    for path, contents in jsonobj.iteritems():
        msg = smap_capnp.SmapMessage.new_message()
        msg.path = path
        msg.uuid = bytes(contents.get('uuid'))
        if contents.get('Contents'):
            msg_contents = msg.init('contents', len(contents.get('Contents')))
            for i,item in enumerate(contents.get('Contents')):
                msg_contents[i] = item
        if contents.get('Readings'):
            msg_readings = msg.init('readings', len(contents.get('Readings')))
            for i, item in enumerate(contents.get('Readings')):
                msg_readings[i] = smap_capnp.SmapMessage.Reading.new_message(time= item[0], data= item[1])
        if contents.get('Properties'):
            msg_properties = msg.init('properties', len(contents.get('Properties')))
            for i, kv in enumerate(contents.get('Properties').iteritems()):
                msg_properties[i] = smap_capnp.SmapMessage.Pair.new_message(key = kv[0], value = kv[1])
        if contents.get('Metadata'):
            md = buildkv('',contents.get('Metadata'))
            msg_metadata = msg.init('metadata', len(md))
            for i, kv in enumerate(md):
                msg_metadata[i] = smap_capnp.SmapMessage.Pair.new_message(key = kv[0], value = kv[1])
        messages.append(msg)
    return messages

def capnp2json(capnpmsg):
    ret = {}
    # tlk = top level key, tlv = top level value
    for tlk,tlv in capnpmsg.to_dict().iteritems():
        # resolve contents as list of strings
        if tlk.lower() == 'contents':
            ret['Contents'] = tlv
        # resolve readings as list of number pairs
        elif tlk.lower() == 'readings':
            ret['Readings'] = map(lambda x: [x['time'], x['data']], tlv)
        # resolve list of {'key': key, 'value': value} dicts
        elif isinstance(tlv, list):
            ret[tlk] = {}
            for d in tlv:
                ret[tlk][d['key']] = d['value']
            ret[tlk] = build_recursive(ret[tlk], suppress=[])
        else:
            ret[tlk] = tlv
    return ret
    
def build_request(jsondata, apikey):
    messages = json2capnp(jsondata)
    req = smap_capnp.Request.new_message()
    req.apikey = apikey
    writeData = req.init('writeData')
    msglist = writeData.init('messages', len(messages))
    for i, item in enumerate(messages):
        msglist[i] = item
    return req

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
# turn string into json obj
jsonobj = json.loads(jsondata)
#print "before"
#print jsonobj
#capnpmsgs = json2capnp(jsonobj)
#for msg in capnpmsgs:
#    recv = capnp2json(msg)
#    print "after"
#    print recv


import socket
IP = "0.0.0.0"
PORT = 8002
s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
req = build_request(jsonobj, apikey)
print req
s.sendto(req.to_bytes(), (IP, PORT))
