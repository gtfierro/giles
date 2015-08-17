## Solidifying the Giles Republish Interface


### Query Subscription

Specified with a full query, which is the generic basis of everything we do.

How exactly this works has been unclear recently, and I've gone back and forth
on it for quite some time, so I think it is useful now to walk through some
solid examples of what I expect to see, and more mportantly, what will be the
most useeful for moving forward. What we definitely do *not* want is something
that is only the easiest to implement -- this system must be useful and ultimately
make everyone's lives easier. Remember that premature optimization is just going to
mess everything up, so implement what you *want* to see and then figure it out later.

Static queries are relatively simple to do, and work well as initialization parameters and the like,
but for aapplicatons that change over time, we want to be able to:
* know when streams of data enter/leave our capture group
    * yes, we have NEW and DEL events that are delivered to subscribers (maybe provide an opt-out?)
    * example:
      {NEW: {uuid1: {metadata, properties, readings}, uuid2: {metadata, properties, readings}}, DEL: [uuid3, uuid4]}
* know what streams of data are in our capture group?
    * up to the client to keep track of this
    * there is a nedge case when a second client subscribes to the same query: do they get notification of all currently
      qualifying streams? or just notification on new ones?
      Maybe the answer is when you initiate a subscription query, it is invoked as though it were a one-off, and the subscriber
      gets an immediate answer, then the actual subscription starts and updates are delivered
* know when the results of a query change. What happens in this case?
    * select * where CLAUSE: every time a captured stream changes, you get the entire message
    * select tag1, tag2 where CLAUSE: every time a captured stream changes one of the captured tags, it will report
      that tag
    * select distinct tag1 where CLAUSE: every time the where clause is affected AND the message includes the distinct tags, reevaluate
      the query and deliver the results
    * select data before now where CLAUSE: this is like the normal republish
    * select data in (range1, range2) where CLAUSE: probably going to need the node-type stuff? we can punt on this for now


Every time the results of the query change, deliver the results of the query.

How can we efficiently tell if the results of the query change?
When a metadata comes in w/ a concerned key, we would re-run the query.
Compare the new results against the old results. Maybe hash old results?


### Self Discussion

What would it look like if I wanted to subscribe to the set of all HVACZones on the 4th floor?

select distinct Metadata/HVACZone where Metadata/Location/Floor = 4;

That's a one-time query.
