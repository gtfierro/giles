---
layout: page
title: Setup
---

* [Installation](#Installation)
* [Running Giles](#Running)
* [Development](#Development)

## <a name="Installation"></a>Installation

Giles requires both an installation of MongoDB and a timeseries database.

Install MongoDB using their [installation instructions](http://docs.mongodb.org/manual/installation/).

For timeseries databases, Giles currently supports either
[ReadingDB](https://github.com/SoftwareDefinedBuildings/readingdb/tree/adaptive)
or [Quasar](https://github.com/SoftwareDefinedBuildings/quasar). Install one of them.

Make a note of what IP/Port Mongo and your timeseries databases are running on.

In the deploy directory, there is a sample [supervisord](http://supervisord.org/) script to help
with deployments.

Giles itself is designed to be easy to install. There are several different ways to do it:

#### Installation From Binary

I make occasional binary releases of Giles, which can be found
[here](https://github.com/gtfierro/giles/releases). Binaries are provided for
Mac OS X, and Linux 32 and 64 bit architectures. Technically, Giles can be
compiled for any platform that is [supported by
Go](https://golang.org/doc/install). Binaries have well defined behavior and
are easy to install, but are infrequently updated and may not have the latest
patches.

#### "Installation" From Dockerfile

In the deploy directory, there is a Dockerfile that should handle building and running giles.
It's not fully finished yet, but it's not too difficult to add references to the Mongo and timeseries
database ports necessary to get it fully running.

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

## <a name="Running"></a>Running Giles

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

TODO:
* administration (api keys)
* trouble shooting
* supervisord how-to

## <a name="Development"></a>Development

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
$ cd archiver ; go generate ; cd .. # OR go generate github.com/gtfierro/giles/archiver
$ go build
$ ./giles -h
```

The `go generate` command is for the YACC-based parser in `archiver/query.y`. You should only need
to run the generation if you changed that file.
