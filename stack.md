---
layout: page
title: The Once And Future Stack
---

The current iteration of the "sMAP stack" has changed since the original conception:

* Timeseries Database: [Berkeley Tree Database (BtrDB)](https://github.com/SoftwareDefinedBuildings/quasar)
* Metadata Database: [MongoDB](https://www.mongodb.org/)
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

If you do not have `apt-get` on your system, you can try [`brew`](http://brew.sh/) for Mac OS X or `yum` for RPM systems.
If you are on Windoze, you are on your own.

* git
* librados-dev
* mongodb
* npm
* nodejs
* nodejs-legacy
* supervisor
* mercurial
* curl
* python-dev
* python-pip
* build-essential

`sudo apt-get install python-dev python-pip build-essential librados-dev git mongodb nodejs nodejs-legacy npm supervisor mercurial curl`


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

Once `nodejs` and `npm` are installed, you will need to install both
[bower](http://bower.io/) and
[react-tools](https://www.npmjs.com/package/react-tools) in a special way so
that they are generally accessible on your system. The other node packages can
be installed "locally".

```bash
$ sudo npm install -g bower react-tools
```

#### Meteor

`curl https://install.meteor.com/ | sh`


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

```bash
$ go get -u -a github.com/SoftwareDefinedBuildings/quasar/qserver
$ go install -a github.com/SoftwareDefinedBuildings/quasar/qserver
$ curl -O https://raw.githubusercontent.com/SoftwareDefinedBuildings/quasar/master/quasar.conf
$ qserver -makedb
```

```ini
# /etc/supervisor/conf.d/quasar.conf
[program:quasar]
command=/home/gabe/go/bin/qserver
autostart=true
autorestart=true
stderr_logfile=/var/log/quasar.err.log
stdout_logfile=/var/log/quasar.out.log
```


## <a name="Giles"></a>Giles

```bash
$ go get -u -a github.com/gtfierro/giles
$ go install -a github.com/gtfierro/giles
$ curl -O https://raw.githubusercontent.com/gtfierro/giles/master/giles.cfg
```

```ini
# /etc/supervisor/conf.d/giles.conf
[program:giles]
command=/home/gabe/go/bin/giles
autostart=true
autorestart=true
stderr_logfile=/var/log/giles.err.log
stdout_logfile=/var/log/giles.out.log
```

## <A Name="Plotter"></a>uPMU Plotter

```bash
$ git clone https://github.com/SoftwareDefinedBuildings/upmu-plotter.git
```

Edit `upmuplot/settings.json` and `upmuplot/client/home.js`

```ini
# /etc/supervisor/conf.d/plotter.conf
[program:plotter]
command=/usr/local/bin/meteor --settings settings.json
directory=/srv/plotter/upmuplot
autostart=true
autorestart=true
stderr_logfile=/var/log/plotter.err.log
stdout_logfile=/var/log/plotter.out.log
```

## <a name="Deckard"></a>Deckard

```bash
$ npm install
$ bower install
$ jsx react_src public/build
```

```ini
# /etc/supervisor/conf.d/deckard.conf
[program:deckard]
command=/usr/bin/npm start
directory=/srv/deckard
autostart=true
autorestart=true
stderr_logfile=/var/log/deckard.err.log
stdout_logfile=/var/log/deckard.out.log
```

