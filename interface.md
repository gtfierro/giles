---
layout: page
title: Interface
---

The sMAP archiver is a streaming storage manager which adds tools for storing
time-series data from sMAP sources, and accessing both historical and real-time
data. It may be used as an interface for developing applications which access
sMAP data, or retrieve data for offline analysis.

The archiver API is available over HTTP as well as several alternative
interfaces (mentioned below)

* [API](#API)
* [Query Language](#querylang)
* [Republish (Pub/Sub)](#republish)
* [Data Publication](#datapub)

## <a name="API"></a>API

The Giles HTTP API offers the following endpoints:

* `/add/<key>`: sMAP messages are POSTed to this URL to archive the metadata and data attached for later access. Data posted here will
  also be forwarded to subscribed clients. A valid API key is needed to post data, unless Giles is configured to ignore API keys
* `/api/query`: sMAP queries POSTed to this URL are evaluated and then returned in the body of the response as a JSON object. Queries should be
  syntactically valid as per the query language specification below
* `/republish`: a sMAP where clause posted to this URL will subscribe that client to the set of streams that match the query clause

The default port for this interface is 8079, though this is configurable.

Giles offers non-HTTP interfaces to make it easier to use the archiver from embedded devices, web services and other sources. These interfaces
currently include

* MsgPack / UDP (almost done)
* MsgPack / TCP (almost done)
* JSON / WebSockets (done)
* CapnProto / UDP (not very done)

These interfaces, while different from the usual HTTP interface (no such thing as a "URL" at layer 4), do their best to provide
analogous functionality. Detailed documentation is forthcoming, but currently the easiest way to adapt the Giles interface to non-HTTP
clients is to write a short bit of middleware to do the protocol translation.

## <a name="querylang"></a>Query Language

The sMAP query language (original formulation can be found [here](http://pythonhosted.org/Smap/en/2.0/archiver.html#query-language)) is a simple,
SQL-like language that allows the user to treat Metadata tags like SQL column names. Giles implements a modern reimplementation with an eye towards
extensibility. The full YACC implementation of the sMAP query language is [here](https://github.com/gtfierro/giles/blob/master/archiver/query.y).
**Aside from sMAP operators, which have yet to be implemented**, the Giles-flavored sMAP query language aims to support the full range of old sMAP
queries, as well as some new features.

To execute queries, query strings can be sent as the body of a POST request to
the query-endpoint on an archiver instance. Over the HTTP interface, this might
look something like (for a local archiver)

```bash
$ curl -XPOST -d "select data before now where Metadata/XYZ=123" http://localhost:8079/api/query
```

In the following snippets of documentation, **bolded words** indicate keywords
that are meant to be typed as-is (e.g. if a query definition starts with
**select**, the actual query string will start with the word `select`). Non-bolded
words will be defined elsewhere.

### Select Query

<p class="message"><b>select</b> selector <b>where</b> where-clause</p>

The basic `select` query retrieves a JSON list of documents that match the provided `where-clause`. Each JSON document
will correspond to a single timeseries stream, and will contain the tags contained in the `selector`. Omitting `where where-clause`
from this query will evaluate the `selector` against all timeseries streams in the database.

#### Selector

A `selector` can be

* a comma-separated list of fully-qualified tags: e.g. `Metadata/Tag1, Properties/UnitofTime, Metadata/Location/Building`.

    Example:

    ```sql
    smap> select uuid, Properties/UnitofTime, Metadata/Floor;
    ```

    returns (`...` indicates more records)

      ```json
    [
        ...
        {
            "Metadata": {
                "Floor": "2"
            },
            "Properties": {
                "UnitofTime": "ms"
            },
            "uuid": "f9aeb8b1-d0aa-5682-9592-110a517293c5"
        },
        {
            "Metadata": {
                "Floor": "1"
            },
            "Properties": {
                "UnitofTime": "s"
            },
            "uuid": "fe1d7301-d92e-573a-ae56-ff7bf2953b0b"
        }
        ...
    ]
      ```

* a **distinct** selector, which takes the form of `distinct <tag>`, and returns a JSON list of all unique values of that tag. A `distinct` selector
  that does not contain one and only one tag is an error.

    Example:

    ```sql
    smap> select distinct Metadata/System
    ```

    returns

    ```json
    [
      "GeneralControl",
      "HVAC",
      "Lighting",
      "Monitoring"
    ]
    ```

* an "everything" selector, designated by `*`. Selecting `*` will return the full document (all tags) for each timeseries stream that matches
  the provided where clause

    Example:

    ```sql
    smap> select *;
    ```

    returns

    ```json
    [
    ...
    {
        "Actuator": {
            "MaxValue": 95,
            "MinValue": 45,
            "Model": "continuous"
        },
        "Metadata": {
            "Building": "IOET",
            "Device": "Thermostat",
            "Driver": "smap.drivers.thermostats.imt550c",
            "Floor": "1",
            "HVACZone": "Invention Lab",
            "Model": "IMT550C",
            "Name": "IOET Class IMT550C Thermostat",
            "Role": "Building HVAC",
            "Site": "d5ed4f6e-a8db-11e4-bd8a-0001c0158419",
            "SourceName": "IOET Class",
            "System": "HVAC",
            "Type": "SP",
            "configured": "True"
        },
        "Path": "/buildinghvac/thermostat0/temp_heat_act",
        "Properties": {
            "ReadingType": "double",
            "Timezone": "America/Los_Angeles",
            "UnitofMeasure": "F",
            "UnitofTime": "s"
        },
        "uuid": "dd57fcd6-7b0b-57dd-9ec0-a952f8e6a117"
    },
    ...
    ]
    ```

#### <a name="where"></a>Where

The `where-clause` describes how to filter the result set. There are several operators you can use:
Tag values should be quoted strings, and tag names should not be quoted. Statements can be grouped using parenthesis.
The `where-clause` construction is used in nearly all sMAP queries, not just `select`-based ones.

| Operator | Description | Usage | Example |
|:--------:| ----------- | ----- | ------  |
|  `=`     | Compare tag values.  | `tagname = "tagval"` | `Metadata/Location/Building = "Soda Hall"` |
|  `like`  | String matching. Use Perl-style regex | `tagname like "pattern"` | `Metadata/Instrument/Manufacturer like "Dent.*"` |
| `has`    | Filters streams that have the provided tag | `has tagname` | `has Metadata/System` |
| `and`    | Logical AND of two queries (on either side) | `where-clause and where-clause` | `has Metadata/System and Properties/UnitofTime = "s"` |
| `or`     | Logical OR of two queries | | |
| `not`    | Inverts a where clause | `not where-clause` | `not Properties/UnitofMeasure = "volts"` |
| `in`     | Matches set intersection on lists of tags | `[list,of,tags] in tagname` | `["zone","temp"] in Metadata/HaystackTags` | 


### Data Query

<p class="message">
<b>select data in</b> (start-reference, end-reference) limit as <b>where</b> where-clause
<br />
<b>select data before</b> reference limit as <b>where</b> where-clause
<br />
<b>select data after</b> reference limit as <b>where</b> where-clause
</p>

You can access stored data from multiple streams by using a data query. Data matching the indicated ranges will be returned for each of the
streams that match the provided `where-clause`.

#### As

The `as` component allows a query to specify what units of time it would like the data returned as. The default is milliseconds, but the user
can specify others (ns, us, ms, s) as per the Unix-compatible notation in the Time Reference table below.

For a sample source, here's the same data point with 4 different units of time. Obviously the resolution is only as good as the underlying source. The
archiver does not add additional time resolution, so if our source published in milliseconds, querying for data as micro- or nanoseconds would not return
more detailed information. The sample source here reported in nanoseconds.

**Also note that the nanosecond representation is returned in scientific notation. This is a known issue and will be fixed in an upcoming release**

```
smap> select data before now as s where uuid = "50e4113d-f58e-468f-b197-8b90a49d42e9";
{
  "Readings": [
    [
    1431290271.0,
    577
    ]
  ],
  "uuid": "50e4113d-f58e-468f-b197-8b90a49d42e9"
}

smap> select data before now as ms where uuid = "50e4113d-f58e-468f-b197-8b90a49d42e9";
{
  "Readings": [
    [
    1431290271944.0,
    577
    ]
  ],
  "uuid": "50e4113d-f58e-468f-b197-8b90a49d42e9"
}

smap> select data before now as us where uuid = "50e4113d-f58e-468f-b197-8b90a49d42e9";
{
  "Readings": [
    [
    1431290271944557.0,
    577
    ]
  ],
  "uuid": "50e4113d-f58e-468f-b197-8b90a49d42e9"
}

smap> select data before now as ns where uuid = "50e4113d-f58e-468f-b197-8b90a49d42e9";
{
  "Readings": [
    [
    1.431290271944557e+18,
    577
    ]
  ],
  "uuid": "50e4113d-f58e-468f-b197-8b90a49d42e9"
}

```


#### Limit

The `limit` is optional, and has two components: **limit** and **streamlimit**. **limit** controls the number of points returned per stream,
and **streamlimit** controls the number of streams returned. For the **before** and **after** queries, **limit** will always be 1, so it only
makes sense to use **streamlimit** in those cases. The exact syntax looks like

<p class="message">
<b>limit</b> number <b>streamlimit</b> number
</p>

where `number` is some positive integer. Both the **limit** and **streamlimit** components are optional and can be specified independently, together
or not at all.

#### Time Reference

Data can be retrieved for some time region using a range query (`in`) or relative to some point in time (`before`, `after`). These reference times
must be a UNIX-style timestamp, the **now** keyword, or a quoted time string.

Time references use the following abbreviations:

| Unit | Abbreviation | Unix support | Conversion to Seconds |
|:----:|:------------:|:-------------:|--------------------- |
| nanoseconds  | ns | yes |1 second = 1e9 nanoseconds |
| microseconds | us | yes |1 second = 1e6 microseconds |
| milliseconds | ms | yes |1 second = 1000 milliseconds |
| seconds | s | yes |1 second = 1 second |
| minutes | m | no  | 1 minute = 60 seconds |
| hours   | h | no  | 1 hour = 60 minutes |
| days    | d | no  | 1 day = 24 hours |

Time reference options:

* Unix-style timestamp: Unix/POSIX/Epoch time is defined as the number of seconds since 00:00:00 1 January 1970, UTC (Coordinated Universal Time), not
counting leap seconds. In Python, the current Unix time (in seconds) can be found with

    ```python
    import time
    # Python actually returns the milliseconds as a decimal,
    # so we use int to coerce to seconds only
    print int(time.time())
    ```

    Giles includes support for Unix-style timestamps in units other than seconds. By suffixing timestamps with one of the unit abbreviations
    specified above (that have Unix support), we can introduce a finer resolution to our data queries. The following timestamps are all equivalent.

    * `1429655468s`
    * `1429655468000ms`
    * `1429655468000000us`
    * `1429655468000000000ns`

    Specifying a timestamp without units will default to seconds.

* The **now** keyword: uses the current local time as perceived by the server. The **now** time can be adjusted using *relative time references*,
  described below.

* Quoted time strings: Giles supports timestrings enclosed in double quotes that adhere to one of the following formats:

  * `1/2/2006`
  * `1/2/2006 03:04:05 PM MST`
  * `1/2/2006 15:04:05 MST`
  * `1-2-2006` rather than `1/2/2006` is also supported

    These time strings follow the [canonical Go reference time](http://golang.org/pkg/time/#Parse), which is defined to be

    ```
    Mon Jan 2 15:04:05 -0700 MST 2006
    ```
* Relative time references: the above time references are *absolute*, meaning that they define a specific point in time. Using relative time references,
  these absolute times can be altered. The most common form of this is specifying offsets of **now**.

    Relative time references in Giles take the form of `number``unit` where `number` is a positive or negative integer and `unit` is one of the
    abbreviations defined in the table above (not limited to those marked with Unix support). Relative time references can be chained.

    For example, to specify 10 minutes before now, we could use `now -10m`. To specify 15 minutes and 30 seconds after midnight March 13th 2010,
    we could use `"3/13/2010" +15m +30s`

#### Examples

Retrieve the last 15 minutes of data for streams `26955ca2-e87b-11e4-af77-0cc47a0f7eea` and `344783b6-e87b-11e4-af77-0cc47a0f7eea`

```bash
smap> select data in (now -15m, now) where uuid = "344783b6-e87b-11e4-af77-0cc47a0f7eea" or uuid = "26955ca2-e87b-11e4-af77-0cc47a0f7eea";
```

Retrieve a week of data for all streams from Soda Hall

```bash
smap> select data in ("1/1/2015", "1/7/2015") where Metadata/Location/Building = "Soda Hall";
```

Retrieve the most recent data point for all temperature sensors

```bash
smap> select data before now where Metadata/Type = "Sensor" and Metadata/Sensor = "Temperature";
```

### Set Query

<p class="message">
<b>set</b> set-list <b>where</b> where-clause
</p>

The `set` command applies tags to a set of streams identified by a where-clause. `set-list` is a comma-separated list
of tag names and values, e.g.

```bash
smap> set Metadata/NewTag = "New Value" where not has Metadata/NewTag
```

Unless Giles is configured to ignore API keys, a `set` command will only apply tags to streams that match the where clause
AND have the same API key as the query invoker.

### Delete Query

<p class="message">
<b>delete</b> tag-list <b>where</b> where-clause
<b>delete where</b> where-clause
</p>

Currently, Giles only supports delete queries on metadata, not timeseries data. A delete query is applied to all documents that match the provided where-clause.
`tag-list` is a comma-separated list of tag names. If provided, the delete query will remove those tags from all matched documents. If `tag-list` is ommitted,
the delete query will remove **every document** that matches the where clause.

Example of removing tags from a set of documents

```bash
smap> delete Metadata/System, Metadata/OtherTag  where Metadata/System = "Botched Value";
```

Example of removing set of documents

```bash
smap> delete where Path like "/oldsensordeployment/.*"
```

## <a name="republish"></a>Republish

Giles provides the ability to get near real-time access to data incoming to Giles. This is called *republish* in sMAP parlance, and is a variation
of *content-based pub-sub*. A client registers a subscription with the archiver using a [where clause](#where). Following subscription, the archiver
will forward all data to the client on streams that match the provided where clause. If the metadata for a stream changes, the set of matching
streams is updated for each related query.

HTTP-based republish is initiated by a client sending a POST request containing a where clause to the `/republish` resource on the archiver. This
connection is kept open by the archiver, and real-time data from the subscription is forwarded to the client for as long as the connection is
kept open.

Here is an example of republish using cURL, subscribing to all temperature sensors (for a local archiver)

```bash
$ curl -XPOST -d "Metadata/Type = 'Sensor' and Metadata/Sensor = 'Temperature'" http://localhost:8079/republish
```

The [Python sMAP library](https://github.com/SoftwareDefinedBuildings/smap/tree/unitoftime/) provides a nice helper class for doing republish from Python.
It uses the Python Twisted library for asynchronous networking support:

```python
from twisted.internet import reactor
from smap.archiver.client import RepublishClient

archiverurl = 'http://localhost:8079'

# called every time we receive a new data point
def callback(uuids, data):
    print 'uuids',uuids
    print 'data',data

query = "Metadata/Type = 'Sensor' and Metadata/Sensor = 'Temperature'"
r = RepublishClient(archiverurl, callback, restrict=query)
r.connect()

reactor.run()
```

Republish is also available over WebSockets. If the WebSocket interface is enabled on Giles, then a client can open up a WebSocket-based subscription
by opening a WebSocket to `ws://localhost:8078/republish` (for a local archiver), and then sending the where clause as a message. Here is an example
in Python


```python
from ws4py.client.threadedclient import WebSocketClient

class DummyClient(WebSocketClient):
    def opened(self):

        self.send("Metadata/Type = 'Sensor' and Metadata/Sensor = 'Temperature'")

    def closed(self, code, reason=None):
        print "Closed down", code, reason

    def received_message(self, m):
        print m

try:
    ws = DummyClient('ws://localhost:8078/republish')
    ws.connect()
    ws.run_forever()
except KeyboardInterrupt:
    ws.close()
```

Obviously, these are just Python-based examples. Being web-based technologies,
it is possible to use any language/library you want (that correctly implements
HTTP or WebSockets) to interface with `republish` (or any other feature
of the archiver).

## <a name="datapub"></a>Data Publication

The majority of data is published to the sMAP archiver through the instantiation and execution of a sMAP driver, but this is not the only
way to send data to the archiver. Indeed, for some of the newer features supported by Giles but not yet by the Python sMAP client library distribution,
alternative methods of data publication are the only way to use some functionality.

Illustrated here are the JSON-versions of sMAP objects, though translations of these exist for other non-JSON/HTTP interfaces.

```javascript
{
    "/sensor0": {                           // At the top level of a sMAP object is the Path
        "Metadata": {                       // Metadata describes attributes of the data source, and is specified
            "Location": {                   //  as nested dictionaries. Here, "Berkeley" is under the key
                "City": "Berkeley"          //  Metadata/Location/City
            },
            "SourceName": "Test Source"     // Metadata/SourceName is how the plotter identifies a stream of data
        },
        "Properties": {                         // Properties describe attributes of the stream. These MUST be
            "Timezone": "America/Los_Angeles",  //  kept consistent, because they affect how the stream is stored.
            "ReadingType": "double",        // If a "numeric" stream, designates the class of number permitted
            "UnitofMeasure": "Watt",        // Units of measure for the stream
            "UnitofTime": "ms",             // Units of time used in the timestamp
            "StreamType": "numeric"         // Describes type of data in Readings: "numeric" or "object"
        },
        "Readings": [                       // This is an array of (timestamp, value) tuples. Timestamps should be
            [                               //  consistent with Properties/UnitofTime
                1351043674000,              // A timestamp
                0                           // A numeric value
            ],
            [                               // Readings can contain more than one tuple
                1351043675000,
                1
            ]
        ],
        "uuid": "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61" // The globally unique identifier for this stream
    }
}
```

Each sMAP object sent to the archiver MUST contain at least the top-level `Path`, which contains a dictionary with the `Readings` and `uuid`
keys, e.g.

```json
{
    "/sensor0": {
        "Readings": [
            [
                1351043674000,
                0
            ]
        ],
        "uuid": "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61"
    }
}
```

Typically, the first object sent to the archiver for a new stream "initializes" the stream by sending all of the Metadata and Properties at once,
and then just sending the minimal object above for updates to the Readings. Metadata/Properties can be changed by including the updates for
those keys/values in the sent object, much like the "initial" object. For example, if we wanted to update the Metadata for the above stream
to change the city from Berkeley to Mendocino, we would send the following object

```json
{
    "/sensor0": {
        "Metadata": {
            "Location": {
                "City": "Mendocino"
            }
        },
        "Readings": [
            [
                1351043674000,
                0
            ]
        ],
        "uuid": "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61"
    }
}
```

For the HTTP interface, each of these JSON objects would be sent as the body of a HTTP POST request sent to the `/add/<key>` resource of a running
archiver.

In Python, using the [`requests`](http://docs.python-requests.org/en/latest/) library, this would look like

```python
import requests
import json
archiverurl = "http://localhost:8079/add/apikey"
smapMsg = {
    "/sensor0": {
        "Readings": [
            [
                1351043674000,
                0
            ]
        ],
        "uuid": "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61"
    }
}
requests.post(archiverurl, data=json.dumps(smapMsg))
```

It is good practice to include the `Content-Type: application/json` HTTP header, though many libraries will add this automatically.

### Publishing Objects

Recently introduced to Giles is the ability to archive and subscribe to
non-numeric data. This should be considered an **extremely alpha** feature, and potentially buggy.

Usage is very similar to the normal numeric interface, with two exceptions.

Firstly, all object-based streams (rather than numeric streams) must have `Properties/StreamType = "object"` set. A stream cannot publish
both objects and numeric data (unless that numeric data is transmitted as an object) because the archiver needs to track which data store
to query data from. Streams should be made numeric-only wherever possible, because this enables a much richer set of operations and queries
upon the data.

Secondly, the `Readings` portion of a sMAP message, instead of only having numbers as the second element of each 2-tuple reading, can now
contain any JSON-serializable data. This means:

* numbers
* strings
* arrays (of any of these data types)
* dictionaries (of any of these data types)

For object storage, Giles encodes each object as a [MsgPack](http://msgpack.org/)-encoded binary string. Giles places no restrictions
on the consistency of objects, so an individual stream is not limited to pushing only arrays or only strings, but can vary the data-type.

Here is an example of a stream that pushes arrays

```javascript
{
    "/sensor0": {
        "Metadata": {
            "Location": {
                "City": "Berkeley"
            },
            "SourceName": "Test Source"
        },
        "Properties": {
            "Timezone": "America/Los_Angeles",
            "ReadingType": "double",
            "UnitofMeasure": "Watt",
            "UnitofTime": "ms",
            "StreamType": "object"          // This denotes this timeseries as an object-stream
        },
        "Readings": [
            [
                1351043674000,              // A timestamp
                [1,2,3]                     // A vector value (JSON serializable)
            ],
            [                               // Readings can still contain more than one object
                1351043675000,
                ["a","b","c"]               // Object types do not have to be consistent
            ]
        ],
        "uuid": "d24325e6-1d7d-11e2-ad69-a7c2fa8dba61" // The globally unique identifier for this stream
    }
}
```
