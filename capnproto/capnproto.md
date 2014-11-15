## Capn Proto for sMAP

This is a bit of a tricky situation, because while there is a common substructure to those messages,
the key/value pairs are not necessarily hard coded.

For instance, upon starting a source, a path metadata message is sent:

```json
{
    "/": {
        "Contents": [
            "fast"
        ]
    },
    "/fast": {
        "Contents": [
            "sensor0",
            "sensor1",
            "sensor2",
            "sensor3",
            "sensor4"
        ]
    },
    "/fast/sensor0": {
        "Properties": {
            "ReadingType": "long",
            "Timezone": "America/Los_Angeles",
            "UnitofMeasure": "V",
            "UnitofTime": "s"
        },
        "Readings": [],
        "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
    },
}
```

and then upon receiving readings afterwards, those follow the following pattern

```json
{
    "/fast/sensor0": {
        "Readings": [
            [
                1415870695,
                4
            ]
        ],
        "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
    }
}
```

Translating these messages into Capn Proto land, we will have to explicitly label the components:

(maybe something like this? definitely need to think about it more..)

```proto
struct reading {
   time @0 :UInt64;
   data @1 :Float64;
}

struct message {
   path @0 :Text;
   uuid @1 :Text;
   readings @2 :List(reading)
}

struct smap {
   messages @0 :List(message);
}
```
