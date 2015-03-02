## Deployments

A collection of files to make deployments easier

* `giles.conf`: a sample [supervisord](http://supervisord.org/) configuration file. Makes these assumptions (though you can of course change them):
    * [`giles.cfg`](https://raw.githubusercontent.com/gtfierro/giles/master/giles.cfg) is at `/etc/giles.cfg`
    * giles is installed in `$GOROOT/bin/giles`, which should be the default if you installed it from `go get`
    * your deployment server has a user named giles which has access to both of the aforementioned paths

* `Dockerfile`: a Dockerfile that contains giles. Does not provide the
  timeseries or metadata stores -- those must be linked.  Still under
  development, but is a good starting point
