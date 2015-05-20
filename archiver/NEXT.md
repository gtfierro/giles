## Move towards io.{Reader, Writer}

This will help take care of sending and receiving large amounts of data. Using Reader and Writers instead of our
bulk reads and writes will not only improve time spent allocating, but will also force a more pipeline oriented
design.

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
