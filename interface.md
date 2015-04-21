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
* Republish (Pub/Sub)
* Data Publication

## <a name="API"></a>API

The Giles HTTP API offers the following endpoints:

* `/add/<key>`: sMAP messages are POSTed to this URL to archive the metadata and data attached for later access. Data posted here will
  also be forwarded to subscribed clients. A valid API key is needed to post data, unless Giles is configured to ignore API keys
* `/api/query`: sMAP queries POSTed to this URL are evaluated and then returned in the body of the response as a JSON object. Queries should be
  syntactically valid as per the query language specification below
* `/republish`: a sMAP where clause posted to this URL will subscribe that client to the set of streams that match the query clause

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

Syntax: **select** selector **where** where-clause

The basic `select` query retrieves a JSON list of documents that match the provided `where-clause`. Each JSON document
will correspond to a single timeseries stream, and will contain the tags contained in the `selector`. Omitting `where where-clause`
from this query will evaluate the `selector` against all timeseries streams in the database.

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

The `where-clause` describes how to filter the result set. There are several operators you can use:
Tag values should be quoted strings, and tag names should not be quoted. Statements can be grouped using parenthesis.
The `where-clause` construction is used in nearly all sMAP queries, not just `select`-based ones.

| Operator | Description | Usage | Example |
|:--------:| ----------- | ----- | ------  |
|  `=`     | Compare tag values.  | `tagname = "tagval"` | `Metadata/Location/Building = "Soda Hall"` |
|  `like`  | String matching. Use `%` to act as a wildcard (think like regex `.*`) | `tagname like "pattern"` | `Metadata/Instrument/Manufacturer like "Dent%"` |
| `has`    | Filters streams that have the provided tag | `has tagname` | `has Metadata/System` |
| `and`    | Logical AND of two queries (on either side) | `where-clause and where-clause` | `has Metadata/System and Properties/UnitofTime = "s"` |
| `or`     | Logical OR of two queries | | |
| `not`    | Inverts a where clause | `not where-clause` | `not Properties/UnitofMeasure = "volts"` |


### Data Query
