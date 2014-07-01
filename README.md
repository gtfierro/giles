#sMAP Archiver

The [sMAP](https://github.com/SoftwareDefinedBuildings/smap) archiver written in Go to better handle concurrency and stability, moving away from Python and Twisted.

Will initially only implement the `/add` resource, with plans to add `/api/query` and the like. This will not change the archiver interface, but rather re-implement
it in a new language to help with scalability and stability.

Archiver: [http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html](http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html)

Archiver API: [http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#archiverapi](http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#archiverapi)

Data Format: [http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#manual-data-publication-json-edition](http://www.cs.berkeley.edu/~stevedh/smap2/archiver.html#manual-data-publication-json-edition)

`pbuf/rdp.proto` is from [readingDB](https://github.com/stevedh/readingdb).
