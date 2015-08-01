## Solidifying the Giles Republish Interface

There are 2 types of subscriptions in Giles:

### Data Subscription:

Specified with a where clause.

Delivers:

```
{
 "Readings": [[time, val],[time, val],...],
 "uuid": uuid
}
```

The set of reporting streams for a subscription changes automatically,
but this is not relayed over thi channel

### Metadata Subscription:

Specified with a where clause.

3 types of messages:
* a new stream qualifies for the clause. The "new" key delivers a list of the full metadata
  for each stream that qualifies
    ```
    {
     "New": [
             {"uuid": uuid,
              "Metadata": {...},
              "Properties": {...},
              "Path": path},
             ...
            ]
    }
    ```
* a stream that qualifies changes, but still qualifies. Full metadata for each stream is delivered
    ```
    {
     "Change": [
             {"uuid": uuid,
              "Metadata": {...},
              "Properties": {...},
              "Path": path},
             ...
            ]
    }
    ```
* a stream changes from qualified to unqualified
    ```
    {
     "Leave": [uuid1, uuid2, uuid3]
    }
    ```

