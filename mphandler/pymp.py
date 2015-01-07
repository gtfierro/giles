import msgpack
import json
import socket

jsondata = """
{
    "Path": "/fast/sensor0",
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
    "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e",
    "key": "pihlHaUYQGcgOleO-l5-fg6-WxyPJw76s4orcrpA0JC_v8r1wxZiWu1ODhklLwcs9BAXs6B0Soaggd3mFcJYVw=="
}
"""
# turn string into json obj
jsonobj = json.loads(jsondata)
sendbytes = msgpack.packb(jsonobj)

s = socket.create_connection(("localhost",8003))
s.send(sendbytes)
