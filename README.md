## Giles

[![GoDoc](https://godoc.org/github.com/gtfierro/giles?status.svg)](https://godoc.org/github.com/gtfierro/giles)

Giles is a replacement sMAP Archiver that offers more mechanisms for extension and scalability.

From the [sMAP documentation](http://pythonhosted.org/Smap/en/2.0/),
<blockquote>

An enormous amount of physical information; that is, information from and about
the world is available today as the cost of communication and instrumentation
has fallen. However, making use of that information is still challenging. The
information is frequently siloed into proprietary systems, available only in
batch, fragmentary, and disorganized. The sMAP project aims to change this by
making available and usable:

* a specification for transmitting physical data and describing its contents,
* a large set of free and open drivers with communicating with devices using
  native protocols and transforming it to the sMAP profile, and tools for
  building, organizing, and querying large repositories of physical data.
</blockquote>

The sMAP Archiver connects to a timeseries database (either
[readingdb](https://github.com/SoftwareDefinedBuildings/readingdb/tree/adaptive)
or [Quasar](https://github.com/SoftwareDefinedBuildings/quasar)) and a metadata
storage (previously [PostgreSQL](http://www.postgresql.org/), and now
temporarily [MongoDB](http://www.mongodb.org/)) and provides a place for sMAP
drivers and instruments to send their data. It supports both historical data
access as well as a limited realtime publish-subscribe interface. Metadata is
used to describe, filter and select streams of data.

What Giles offers above the original sMAP archiver implementation is the
ability to switch between backend databases and an increased flexibility in the
permitted interfaces/protocols for data. Rather than only supporting JSON/HTTP,
Giles allows data to be sent and received over MsgPack/UDP, MsgPack/TCP,
ProtoBuf/UDP, CapnProto/TCP and JSON/Websockets. It is also very easy to add a
new interface adapter.

There are some cool new features of Giles under active formulation and
development that I won't mention here, but should hopefully be seen soon!

### Install

You will need go version >= 1.4.

```bash
go get -u -a github.com/gtfierro/giles
pip install gilescmd
```

You can now run the `giles` comand. You can see the usage with `giles -h`.

Documentation is available at http://godoc.org/github.com/gtfierro/giles 

#### Installing from Source

If installing from source, clone the giles git repo and then install the go dependencies:

```
$ git clone https://github.com/gtfierro/giles
$ cd giles && ./install_deps.sh
```

For development, I either work in `$GOPATH/src/github.com/gtfierro/giles/...`, which is the default
path where the giles libs are installed, or I will sym link the git repo to there:

```
ln -s path/to/giles/repo/root $GOPATH/src/github.com/gtfierro/giles
```

should take care of it. Now you should be able to compile giles by running

```
$ cd path/to/giles/repo/root
$ cd giles
$ go build
$ ./giles -h
```
