![](http://sufr.kyleterry.com/static/images/sufr-logo.svg)
####
SUFR (Simple URL and Fact Repository) is the successor to shitbucket. In short,
it is a self-hosted URL bookmarker. Checkout my personal copy http://sufr.kyleterry.com/ for an idea of what it looks like. I upload new builds all the time.

There are some [screenshots](./screenshots) in the repo.

## Install
* Clone or `go get github.com/kyleterry/sufr` into your `$GOPATH`
* `cd ${GOPATH}/src/github.com/kyleterry/sufr`
* `make`
* `sudo make install`

This will install `sufr` into the default `PREFIX` which is `/usr/local/bin/`. Omit the last install step if you want to copy the binary somewhere else.

NOTE: I will be crosscompiling binaries. Everything (assets and templates and
database) are compiled into SUFR so you will only need the one binary to run it.
There is no need to copy css, html, and javascript around.

## Running
Once you install SUFR, you can simply run the binary and access the address in
the browser. You will see a setup page and will be able to configure your
instance from there.

When you run the
program, it will generate a directory in `${HOME}/.config/sufr/data` for it's
database file. This can be backed up from the settings page.

### Running in Docker

There is a Docker image available on Docker hub:
`docker pull kyleterry/sufr`  
`docker run -P -d --name sufr kyleterry/sufr`

## Dev mode
SUFR has a `-debug` flag that doesn't currently do much. It just starts a goroutine to spit out database stats every 10 seconds. I will add better debugging in the near future.
