## Giles Integration Tests

Going to define an input file for integration tests that specify the
inputs/outputs for a certain transaction/interaction. The executor will take
each file and run it, connecting clients, etc as needed.


### Examples

```
[Test]
Name: Add data to stream
Layout: Input:1 -> Output:1

[Input:1]
Interface: HTTP
Method: POST
URI: http://localhost:8079/add/apikey
Format: JSON
Data: {'/path': {'Readings': ...}}

[Output:1]
Interface: HTTP
Code: 200
Contents: ''
Format: string
```

The `[test]` section defines the name of the interaction. The input and output
sections take an identifier that specifies which client the input/output are
for. Here, because Input and Output both define client `1`,  a single client
will be instantiated that takes our input and expects the given output. We
could also define an experiment where we need to coordinate 2 clients with
input on the first and expected output on the second.

`Format` defines how we should serialize the data before using it in the input,
or how to deserialize when we receive in the output.

`Layout` specifies the order of client execution. In the trivial case, the test starts Input:1, and then
waits for the result in Output:1.

---

```
[Test]
Name: Query data from stream
Layout: Input:1 -> Output:1

[Input:1]
Interface: HTTP
Method: POST
URI: http://localhost:8079/api/query
Format: string
Data: 'select data in (now -5min, now) where uuid = "XYZ"'
Sleep: 5s

[Output:1]
Interface: HTTP
Code: 200
Contents: {'Readings': [[0, 1], [0,2]]...}
Format: JSON
```

Here, we sleep for 5 seconds before sending the test query. 

---

```
[Test]
Name: Republish
Layout: Input:1 -> Output:1; Input:2 -> Output:2

[Input:1]
Interface: HTTP
Method: POST
URI: http://localhost:8079/add/apikey
Format: JSON
Data: {'/path': {'Metadata'..., 'Readings'...}}

[Output:1]
Interface: HTTP
Code: 200
Contents: ''
Format: string

[Input:2]
Interface: HTTP
Method: POST
URI: http://localhost:8079/republish
Format: string
Data: "Metadata/Tag = 'XYZ'"

[Output:2]
Interface: HTTP
Code: 200
Contents: {'Readings': [[....]], 'uuid': ...}
Format: JSON
```

In `Layout`, the semicolon specifies that the clients should be run in parallel and the test terminates
when both Output:1 and Output:2 have been received.

---

```
[Test]
Name: Query data from stream after adding
Layout: Input:1 -> Output:1 -> Input:2 -> Output:2

[Input:1]
Interface: HTTP
Method: POST
URI: http://localhost:8079/add/apikey
Format: JSON
Data: {'/path': {'Readings': ...}}

[Output:1]
Interface: HTTP
Code: 200
Contents: ''
Format: string

[Input:2]
Interface: HTTP
Method: POST
URI: http://localhost:8079/api/query
Format: string
Data: 'select data in (now -5min, now) where uuid = "XYZ"'
Sleep: 5s

[Output:2]
Interface: HTTP
Code: 200
Contents: {'Readings': [[0, 1], [0,2]]...}
Format: JSON
```

`Layout` here tells us to only start client Input:2 after Output:1 has been
received. Additionally, once Output:1 has been received, Input:2 tells us to
sleep for 5 seconds prior to starting.
