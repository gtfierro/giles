## Giles

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
