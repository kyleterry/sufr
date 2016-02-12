# SUFR
SUFR (Simple URL and Fact Repository) is the successor to shitbucket. In short,
it is a self-hosted URL bookmarker.

## Install
* Clone or `go get github.com/kyleterry/sufr` into your `$GOPATH`
* `cd ${GOPATH}/src/github.com/kyleterry/sufr`
* `make`
* `sudo make install`

NOTE: I will be crosscompiling binaries. Everything (assets and templates and
database) are compiled into sufr so you will only need the one binary to run it.
There is no need to copy css, html, and javascript around. When you run the
program, it will generate a director in `${HOME}/.config/sufr/data` for it's
database file. This can be backed up from the settings page.

## Running
Once you install SUFR, you can simply run the binary and access the address in
the browser. You will see a setup page and will be able to configure your
instance from there.
