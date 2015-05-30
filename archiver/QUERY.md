## Query Operators

Planning on supporting:

* `group by`: create nested lists of streams, grouped by some shared attribute
* `order by`: order the collection of streams, either explicitly (e.g. 'uuid1, uuid2, uuid3')
  or by some other means (alphabetically, max value, etc)
* `max`,`min`,`count`,`mean`,`median`,`mode`
* `zip`, `align`: line up to timeseries by their timestamps. This will require some algorithm for
  doing interpolation or filling or sampling
* `join`: joins two or more timeseries into a single stream. This can be a "fill", where one timeseries
  fills in the gaps of another, or some sort of merge where they are added or subtracted. This can optionally
  be combined with `zip` (as it will be in the sum/subtract cases, probably)
* `window`: aggregate timeseries to time windows. These can be sliding or discrete. Needs an algorithm
  for how to compute windows (mean, max, min, sum, etc)
* `rate`: get the average report rate of a stream over some window
* `edge`: stream of the differences between each point and the previous point

Things to get working soon:
window
align
external operators

### Considerations

* update interval: are these queries updated every time new data is published to one of the concerned streams?
  This is probably reasonable to do, especially if we are able to merge stream operations in the graph
  to avoid duplicate computation/checking
* consistency:
    * axis: 0 = columnwise, 1 = rowwise

## InterNode communication

The problem with moving ahead with the operator implementations is that the interface between nodes is not quite clear.

There are several types of data that can come through a node:
* timeseries of objects
* timeseries of numbers
* a list of streamIDs
* a generic document

A timeseries has 2 attributes:
* a UUID (uniquely identifies a timeseries)
* a list of readings (each reading is a tuple of [timestamp, value])

Let's walk through a couple examples and build up what happens:

```
apply min(axis=0) to data in (now -1m, now) where Metadata/XYZ=123;
```

The where clause forms a `whereNode`. The output of a wherenode is a list of
UUIDs, and this is fed to the input of the `selectNode`. 

The select node ("data in (now -1m, now)") applies the data selector to the
list of UUIDs it is given. The output of the select node is a timeseries,
which can either be a timeseries of objects or of numbers (or possibly both,
if the UUIDs are mixed...we probably want to error out in this case). For
now, let's assume that 

The answer is to have 1 node for each operation, e.g. min, groupby ,etc. This node defines
the mappings of input to output for each possible input type. This way, during the formation of the
tree, we can check the output/input combinations at runtime. No need to implement a completely different
node for each specific type of input. 
We should also have tags that identify the input/output type of a node, and then each node is able to parse
the input/output types of the node as it wants. For example, I might have an "order by" node that can act
on timeseries of either objects or numbers, so it only needs to check if the input is a "timeseries" rather
than doing a bigger OR clause.

So, what is the set of tags?

First, list all the different types of input/output you can imagine

list of timeseries
    select data in (now -1m, now) where Metadata/tag = value // returns multiple timeseries in an array

timeseries of objects
    select data in (now -1m, now) where uuid = unique val // a single timeseries of objects

timeseries of scalars
    select data in (now -1m, now) where uuid = unique val // a single timeseries of scalars

a list of UUIDs
    `select uuid where Metadata/tag = true`
    This actually probably gives us a list of documents. Maybe we could have a 'flatten' command, or an 'extract' command
    to return a single field from each document?

a metadata document
    `select * where uuid = uniqueid`
    Returns us the document, but I don't know where this would actually be used as input or output? It would just be a "normal"
    select clause. No need for operators

a list of metadata documents
    `select * where metadata/tag = someval`
    Same as above, I don't think we support operators on these

a list of scalars
    `apply min(axis=0) to data in (now -1m, now) where metadata/tag = value`
    `apply min(axis=0) to data in (now -1m, now) where uuid = uniqueval`
    The first is obviously a list of scalars that come from the metadata evaluation, but what does
    the result look like? A straight list probably wouldn't be that helpful, so the intermediate
    value should probably be a list of uuid structs, e.g.
    ```
    [{'uuid': 'uuid1', 'result': 30},
     {'uuid': 'uuid2', 'result': -32}]
    ```
    Then, a user could apply an 'extract' operator to just retrieve 'result' if they wanted.
    The second example will only return a single value, but I don't think it necessarily makes
    sense to treat the 'one' case as different from a metadata clause that just returns 
    a single document. It should be an array either way, so this doesn't end up being a special
    case for every single node.

a list of objects
    While we probably won't support any direct operations on timeseries of objects, we need to account
    for this data type because a user could have an external operator that takes in a list of timeseries
    of objects and then performs some special operation and returns these. As with the list of scalars,
    the output type of this should be a list of uuid structs:
    ```
    [{'uuid': 'uuid1', 'result': [0,2,3]},
     {'uuid': 'uuid2', 'result': "stringval"}]
    ```

`apply orderby(uuid1, uuid2, uuid3) to data in (now -1m, now) where ....`


Do we have a better idea of tags now:
* structure: list, timeseries (mutually exclusive)
* datatype: scalar, object (mutually exclusive)

This is probably good enough for now.

Things to change:
nodes no longer need kv argument. Tags will be handled internally by the node. How do we organize, though?
We are dealing with an interface, so a method accessor would be best.
What output structures do you support?
Do you support input structure X?
Do you support input structure + data type X Y ?
What inputs datatypes, structures do you accept?

Internally:

map[string]uint = {
"out:structure" = LIST
"out:datatype" = SCALAR
"in:structure" = LIST
"in:datatype" = SCALAR
}

structure: list 0 timeseries 1
datatype:  obj  0 scalar 1

Node: list (obj, scalar)
Node: timeseries scalar

check first the structure. if structure doesn't match, we fail.
Then check datatype. If datatype not specified on Input, we win.
if datatype don't match, we fail.


API:

// either can be nil, and it will not check that part
HasOutput(structure StructureType, datatype DataType) bool

HasInput(structure StructureType, datatype DataType) bool

So when I'm connecting nodes, how do I know what the inputs and outputs are? I
know what kind I accept.  When nodes are connected, we check that the output of
the parent "fits into" the input of the child.



