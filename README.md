## Giles

[![GoDoc](https://godoc.org/github.com/gtfierro/giles?status.svg)](https://godoc.org/github.com/gtfierro/giles)

Documentation is available at http://godoc.org/github.com/gtfierro/giles 

Giles is a replacement sMAP Archiver that offers more mechanisms for extension and scalability.

From the [sMAP documentation](http://pythonhosted.org/Smap/en/2.0/),
<blockquote>

An enormous amount of physical information; that is, information from and about
the world is available today as the cost of communication and instrumentation
has fallen. However, making use of that information is still challenging. The
information is frequently siloed into proprietary systems, available only in
batch, fragmentary, and disorganized. The sMAP project aims to change this by
making available and usable:

<ul>
<li>a specification for transmitting physical data and describing its contents,</li>
<li>a large set of free and open drivers with communicating with devices using
  native protocols and transforming it to the sMAP profile, and tools for
  building, organizing, and querying large repositories of physical data.</li>
</ul>
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

### Installation

Giles requires both an installation of MongoDB and a timeseries database.

Install MongoDB using their [installation instructions](http://docs.mongodb.org/manual/installation/).

For timeseries databases, Giles currently supports either
[ReadingDB](https://github.com/SoftwareDefinedBuildings/readingdb/tree/adaptive)
or [Quasar](https://github.com/SoftwareDefinedBuildings/quasar). Install one of them.

Make a note of what IP/Port Mongo and your timeseries databases are running on.

Giles itself is designed to be easy to install. There are several different ways to do it:

#### Installation From Binary

I make occasional binary releases of Giles, which can be found
[here](https://github.com/gtfierro/giles/releases). Binaries are provided for
Mac OS X, and Linux 32 and 64 bit architectures. Technically, Giles can be
compiled for any platform that is [supported by
Go](https://golang.org/doc/install). Binaries have well defined behavior and
are easy to install, but are infrequently updated and may not have the latest
patches.

#### Installation From Source

You will need Go version >= 1.4, which you can install from the [official Go page](https://golang.org/doc/install).
Make sure that your `$GOPATH` environment variable is configured correctly and is on your `$PATH`. It is also
a good idea to add `$GOPATH/bin` to the end of your `$PATH` as well.

Giles also requires [Mercurial](http://mercurial.selenic.com/downloads) to be installed to fetch some packages.

To retrieve the giles source code and install the executable, run

```bash
$ go get -u -a github.com/gtfierro/giles # fetches source
$ go install -a github.com/gtfierro/giles # compiles and moves executable into $PATH
```

### Running Giles

You should now be able to run the `giles` comand. You can see the usage with `giles -h`.

Giles requires knowledge of a configuration file in order to run. The default
configuration file can be downloaded from
[here](https://raw.githubusercontent.com/gtfierro/giles/master/giles.cfg), and
should be mostly self explanatory. Remember to alter the TSDB, ReadingDB/Quasar and Mongo
sections to match your deployment.

You can now run giles and see the following output (or something similar to it)

```bash
$ giles -c path/to/giles.cfg
Giles Configuration
Connecting to Mongo at 0.0.0.0 : 27017
Using Timeseries DB readingdb
        at address 0.0.0.0 : 4242
        with keepalive 30
Profiling disabled
NOTICE Feb 25 16:23:33 metadata.go:35 ▶ Connecting to MongoDB at 0.0.0.0:27017...
NOTICE Feb 25 16:23:33 metadata.go:41 ▶ ...connected!
NOTICE Feb 25 16:23:33 readingdb.go:100 ▶ Connecting to ReadingDB at 0.0.0.0:4242...
NOTICE Feb 25 16:23:33 readingdb.go:101 ▶ ...connected!
ERROR Feb 25 16:23:33 ssshscs.go:127 ▶ Failed to open authorized_keys file (open /home/gabe/.ssh/authorized_keys: no such file or directory)
INFO Feb 25 16:23:33 ssshscs.go:103 ▶ Listening on 2222...
NOTICE Feb 25 16:23:33 handlers.go:46 ▶ Starting HTTP on 0.0.0.0:8079
NOTICE Feb 25 16:23:33 handlers.go:28 ▶ Starting CapnProto on 0.0.0.0:1235
NOTICE Feb 25 16:23:33 handlers.go:46 ▶ Starting WebSockets on 0.0.0.0:1234
NOTICE Feb 25 16:23:33 handlers.go:82 ▶ Starting MsgPack on UDP [::]:1236
INFO Feb 25 16:23:33 stats.go:42 ▶ Repub clients:0--Recv Adds:0--Pend Write:0--Live Conn:0
NOTICE Feb 25 16:23:33 handlers.go:60 ▶ Starting MsgPack on TCP 0.0.0.0:1236
INFO Feb 25 16:23:34 stats.go:42 ▶ Repub clients:0--Recv Adds:0--Pend Write:0--Live Conn:0
INFO Feb 25 16:23:35 stats.go:42 ▶ Repub clients:0--Recv Adds:0--Pend Write:0--Live Conn:0
```

### Development

For development, I either work in `$GOPATH/src/github.com/gtfierro/giles/...`, which is the default
path where the giles libs are installed, or I will sym link the git repo to there:

```
ln -s path/to/giles/repo/root $GOPATH/src/github.com/gtfierro/giles
```

should take care of it. Now you should be able to compile giles by running

```bash
$ cd path/to/giles/repo/root
$ cd giles
$ go get ...
$ go build
$ ./giles -h
```
