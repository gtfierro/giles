## Giles

Given that I am planning on dramatically expanding the range of incoming
interfaces that Giles supports, there should really be a core library of
functionality, and the various interfaces should just translate the incoming
data into a form understood by that library. This will help avoid repeated
code.

So, what are those functions?

timeseries:

* add data (list of points) -> success=t/f
* get data (list of ids, start, end) -> list of data, error
* prev data (list of ids, start, limit) -> list of data, error
* next data (list of ids, start, limit) -> list of data, error

metadata:

* get tags (select tags, where tags) -> tag collection, error
* get uuids (where tags) -> list of uuids, error
* set tags (update tags, where tags) -> num changed, error

```go
AddData(readings []interface{}) -> bool
GetData(streamids []string, start, end uint64) -> ([]interface{}, error)
PrevData(streamids []string, start uint64, limit int32) -> ([]interface{}, error)
NextData(streamids []string, start uint64, limit int32) -> ([]interface{}, error)
GetTags(select_tags, where_tags map[string]interface{}) -> (map[string]interface{}, error)
GetUUIDs(where_tags map[string]interface{}) -> ([]string, error)
SetTags(update_tags, where_tags map[string]interface{}) -> (int, error)
```
