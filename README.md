![](http://sufr.kyleterry.com/static/images/sufr-logo.svg)
####
SUFR (Simple URL and Fact Repository) is the successor to shitbucket. In short,
it is a self-hosted URL bookmarker. Checkout my personal copy http://sufr.kyleterry.com/ for an idea of what it looks like. I upload new builds all the time.

There are some [screenshots](./screenshots) in the repo.

[![Build Status](https://travis-ci.org/kyleterry/sufr.svg?branch=master)](https://travis-ci.org/kyleterry/sufr)

## Features

* Save links, videos and image galleries and access them from any internet
  connected device with a browser
* Private instances - make your entire SUFR private and only accessible to you
  if you are logged in
* Private links - sometimes you just don't want people to see what you saved
  even if your instance if public, so you can flag a link as private
* Embed YouTube videos and Imgur galleries/images
* Backup your database with a single link in settings (can be hit via curl in a
  cron job)
* Pin your most used tags to the sidebar

I'm working on some features that are definitely important, you can find them [in my github issues](https://github.com/kyleterry/sufr/issues?utf8=%E2%9C%93&q=is%3Aissue%20is%3Aopen%20label%3Afeature%20).

## Install
* Clone or `go get github.com/kyleterry/sufr` into your `$GOPATH`
* `cd ${GOPATH}/src/github.com/kyleterry/sufr`
* `make`
* `sudo make install`

This will install `sufr` into the default `PREFIX` which is `/usr/local/bin/`. Omit the last install step if you want to copy the binary somewhere else.

NOTE: I will be crosscompiling binaries. Everything (assets and templates and
database) are compiled into sufr so you will only need the one binary to run it.
There is no need to copy css, html, and javascript around.

## Running
Once you install sufr, you can simply run the binary and access the address in
the browser. You will see a setup page and will be able to configure your
instance from there.

When you run the
program, it will generate a directory in `${HOME}/.config/sufr/data` for it's
database file. This can be backed up from the settings page.

### Running in Docker
There is a Docker image available on Docker hub:

`docker pull kyleterry/sufr`  
`docker run -P -d --name sufr kyleterry/sufr`

## Backups
The sufr database can be backed up in the settings. From the UI, click on your username dropdown in the menu bar and then click on `settings`. From the settings page, scroll down to the bottom and look for a link called `Backup Database`. When you click this, it will start a download. This is your database file. It's a binary blob. Keep it safe in case you need to restore.

### Restoring
If you need to restore, copy the database file into `${HOME}/.config/sufr/data/sufr.db` on the machine that sufr runs on.

## Dev mode
sufr has a `-debug` flag that doesn't currently do much. It just starts a goroutine to spit out database stats every 10 seconds. I will add better debugging in the near future.
