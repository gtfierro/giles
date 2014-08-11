#sMAP Archiver

The [sMAP](https://github.com/SoftwareDefinedBuildings/smap) archiver written
in Go to better handle concurrency and stability, moving away from Python and
Twisted.

Will initially only implement the `/add` resource, with plans to add
`/api/query` and the like. This will not change the archiver interface, but
rather re-implement it in a new language to help with scalability and
stability.

Archiver:
[http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html](http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html)

Archiver API:
[http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#archiverapi](http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#archiverapi)

Data Format:
[http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#manual-data-publication-json-edition](http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#manual-data-publication-json-edition)

`sr/archiver/proto/rdb.proto` is from [readingDB](https://github.com/stevedh/readingdb).


## Archiver Interfaces

* `/republish`: if GET request, looks like it just forwards all readings to a
  new URI. If POST request, there is an attached query that returns a list of
  UUIDs. The client will only want those UUIDs in this case. In either case,
  the client will long-poll the `/republish` resource in order to see data as
  it is pushed to the main archiver.

* `/add`: Data is POSTed to `/add/[api key]`; this resource checks that it is a
  valid key and then inserts it into the postgres and readingdb databases. We
  have the readingdb part down, but we will need to combine this with the
  postgres store -- possibly move to mongodb here.  Also need to keep track of
  streamids as they are created.

* `/api`: We will want to preserve this piece of code as much as possible,
  because it handles the building-up of queries for sMAP UUIDs and data. Adjust
  this to talk to the new mongodb instead of the old postgres instance, and that
  should take care of most of the functionality

Another piece that is missing will be the creating of new api keys and the
management of streams. This was previously done through powerdb2, but there
should be a simplified API, simplified web interface and a command-line utility
for creating new stream ids.

The command-line utility should read off some credentials from an internal
config file or environment vars (think AWS creds) and use those as
authentication to create new subscription keys.

### /add

* **Create stream ids if they do not exist**: 
    
* Look up stream ids
* Add data to readingdb
