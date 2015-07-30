---
layout: page
title: The Once And Future Stack
---

The current iteration of the "sMAP stack" has changed since the original conception:

* Timeseries Database: [Berkeley Tree Database (BtrDB)](https://github.com/SoftwareDefinedBuildings/quasar)
* Metadata Database: [MongoDB](https://www.mongodb.org/)
    * also used by BtrDB
* Archiver: [Giles](https://github.com/gtfierro/giles)
* Plotter: [uPMU Plotter](https://github.com/SoftwareDefinedBuildings/upmu-plotter)
* Status Dashboard: [Deckard](https://github.com/gtfierro/deckard)

Though the installation/setup instructions for all of these exist in some form across many links, this page
will bring them all together for a **single definitive installation document** for The Once And Future Stack.

This installation assumes a Debian-based distro such as Ubuntu. These instructions have been developed and tested
on Ubuntu 14.04, 14.10.x and 15.04, but installing these packages on other systems should be straightforward.

## <a name="BasePackages"></a>Base Packages

These are the required packages on the system for the rest of the instructions to work.

### Apt Packages

**[[In Progress]]**

If you do not have `apt-get` on your system, you can try [`brew`](http://brew.sh/) for Mac OS X or `yum` for RPM systems.
If you are on Windoze, you are on your own.

* git
* librados-dev
* mongodb
* npm
* nodejs (see below)
* supervisor
* mercurial

`sudo apt-get install -y librados-dev git mongodb nodejs npm supervisor mercurial`


### Others

#### Go

There are several Go-based components. Occasional binary releases are available
for these, but it is recommended to compile them from source while still under
active development. It is recommended to follow the [official installation
instructions](https://golang.org/dl/), including setting up your `$GOROOT` environment variable.

I prefer to place all environment variables in my `~/.bashrc` file.

Make sure that your `$GOPATH` environment variable is configured correctly and
is on your `$PATH`. Also, add `$GOPATH/bin` to the end of your `$PATH` as well.

This is how I do it:

```
$ mkdir $HOME/go
# inside .bashrc
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

#### NodeJS

You will need to symlink the `nodejs` binary to `/usr/bin/node`, e.g. with ```sudo ln -s `which nodejs` /usr/bin/node```.

Once `nodejs` and `npm` are installed, you will need to install both
[bower](http://bower.io/) and
[react-tools](https://www.npmjs.com/package/react-tools) in a special way so
that they are generally accessible on your system. The other node packages can
be installed "locally".

```bash
$ sudo npm install -g bower react-tools
```

## <a name="MongoDB"></a>MongoDB

Mongo will have been installed by the above aptitude command. For deployments, it is recommended to use the
Mongo service handler, which will handle everything for you:

```bash
$ sudo service mongodb start
```

For development, it can be helpful to run MongoDB in the foreground:

```bash
$ mkdir mongodb_data
$ mongod --dbpath mongodb_data
```

Be aware that if this crashes, you will need to manually restart.

## <a name="BtrDB"></a>BtrDB

