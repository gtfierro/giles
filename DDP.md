Problem: We want to execute metadata queries against the sMAP archiver; the documents representing timeseries
returned from these queries are then rendered in the meteor application. This is fine for the case when we have
a singular identifier that represents some value, e.g. the value for the UUID of a timeseries. The UUID is
not expected to change, so we can use the same identifier for the same stream of data (though discovering these
UUIDs is another matter).

A problem arises when we attempt to use this same avenue for metadata. Here is an example tool flow:

1. Need to render the zone air temperatures for some view in the web
   application

2. Lookup this data in the MongoDB -- turns out we don't have it, so we need to
   go ask the sMAP archiver. In the meantime, we return some null objects that obey the same
   schema as the points we expect; this way Meteor can still render the page

3. Execute some metadata query to retrieve the tags for all temperature sensors
   for HVAC Zone = '410 Soda'. This is returned as a list of JSON objects

4. The list of resulting JSON objects are obtained in a callback off of the HTTP query to the
    sMAP archiver. The data needs to be parsed and structured so that it can be inserted into
    the MongoDB in such a way that it can be reused later.


Right now, MongoDB is essentially duplicating the work of the archiver.

However, the client-server consistency model is very nice -- not a whole lot of motivation to replace
that. Meteor.subscribe might actually include the information into the local client mongos?

Altering or deleting objects or metadata does not make use of this consistency
model, though, and in fact works against it. The meteor application allows the
user to change metadata, and that metadata must be updated both in the mongo
application as well as sent to the sMAP archiver. 

