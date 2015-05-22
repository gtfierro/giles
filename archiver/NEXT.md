## Move towards io.{Reader, Writer}

This will help take care of sending and receiving large amounts of data. Using Reader and Writers instead of our
bulk reads and writes will not only improve time spent allocating, but will also force a more pipeline oriented
design.

### What's Next

Start introducing this to the current interfaces between giles modules. It is already a problem that serialized
JSON byte arrays are being thrown around like they are so cheap. Need to adjust that, think of a general data
model that will take care of all the different sorts of data we are sending (even if this is "just" a msgpack
byte array rather than json), and the n convert the interfaces to use that. This may still be a byte stream,
but at the very least, we can use a Reader/Writer to make the transport more efficient.

FINALLY you MUST benchmark the msgpack interface

## Giles Pipeline

If we are moving more towards a data pipeline that includes operators, those
operators should be able to function on Metadata and objects as well as
timeseries data. The current implementation of dynamic subscriptions hints at what should really happen
later, but doesn't quite get there.

The key idea is to encapsulate the dependencies of the graph process.
The where clause is probably the base of this, and that resolves to streams that can either deliver
objects or timeseries or metadata. For the current republish client, this pipeline is very simple

```
Where query --> set of uuids --> selector --> client
```

but if we want to look into streaming queries, what if we dramatically expanded that?

```
select Metadata/Room where Metadata/Sensor/Type = "Temperatue";
diff Metadata/Room where Metadata/Sensor/Type = "Temperatue"; < ??
```
As a streaming query, this would deliver to the subscriber the changes in the `Metadata/Room`
tag for all streams that are tagged as `metadata/Sensor/Type`.

We can imagine certain other keywords:

* notify me when the set of qualifying streams change:
    ```
    subscribe set Metadata/Room where Metadata/Sensor/Type = "Temperature";

    > {'added': ['uuid1','uuid2'], 'removed': ['uuid3']}
    ```

We do not want to duplicate the whole pipeline for each subsccriber. Figure out what data each client wants
and then build a graph out of that. When we add operators later (not too much later), that is "simply"
a graph that extends that prior model. 

### What's Next

Proof of concept of graph data pipeline. Doesn't have to be performant, just explore how to do it in the code.

## Interfaces

Devices should implement interfaces, e.g. 'thermostat interface', much like TinyOS. it is the burden of the driver
to map the functionality of the device onto the functionality expected by the interface. Once we have a well-defined
series of interfaces, you could "cast" one interface to another, e.g. a camera interface AS an occupancy sensor.
These mappings would probably be known in advance, but should be lightweight, or use sMAP operator transformations
to do trivial translations between items. 

You could have inheritance between items? some fancy thermostat interface that inherits from other interfaces?

## More Operators

Could you have a sMAP operator that pipes to a remote service and gets output?

```
apply mean < post http://myserver:4000/transformdata < window(5min, mean) < data in (now -60min, now) where uuid = 'abcdef'
```

you could distribute computation! fancy python stuff we can't do in go? post to external resource. Maybe you have a folder
of python functions (or whatever language) and there's some wrapper that puts those services written in X language that redirects
POST data into that function, and then redirects STDOUT back to you.

could be an arbitrary URL, OR could be a known keyword that maps to an external operator

makes tony happy! Write stuff in Julia, then the server handles it all for you. The supported operators could be a GIT repository.
Could combine this with DISTIL-like scheduling of processes

This means that giles operators do not have to implement everything under the sun, but everything under the sun IS integratable
without needing a heavy computationally-capable client.

How does this differ from services? They are ad-hoc, and does not require being set up before hand. Computation can be done on the fly.
if we are able to incorporate the set of orators applied to a stream/query, then these could be 'baked in' to sMAp streams
and used as part of the queyr planner to avoid duplicate computation between clients or between different runs of the same client.

## Building Profile

relational database that stores stuff not linked to streams. What are the rooms in a building? which upmus belong to company x?
This can't be associated with a stream directly, so it needs to go somewhere else. BUT it would be nice to have a way to query
it from the same location. Again, no need to innovate here -- just need a way to direct a query between the stream metadata
and the "system" metadata (need better phrase there). OpenBAs *and* uPMU people need this.
