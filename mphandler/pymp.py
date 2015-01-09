import msgpack
import time
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
    "key": "jgkiXElqZwAIItiOruwjv87EjDbKpng2OocC1TjVbo4jeZ61QBqvE5eHQ5AvsSsNO-v9DunHlhjwJWd9npo_RA=="
}
"""
# turn string into json obj
jsonobj = json.loads(jsondata)
sendbytes = msgpack.packb(jsonobj)

s = socket.create_connection(("localhost",8003))
for i in range(10):
    time.sleep(.5)
    s.send(sendbytes)

