# scss-powertools :zap:
[![npm](https://img.shields.io/npm/v/scss-powertools.svg)](https://www.npmjs.com/package/scss-powertools)
[![Greenkeeper badge](https://badges.greenkeeper.io/Tutrox/scss-powertools.svg)](https://greenkeeper.io/)
[![Build Status](https://travis-ci.org/Tutrox/scss-powertools.svg?branch=master)](https://travis-ci.org/Tutrox/scss-powertools)

Lint, compile, prefix and minify¹ SCSS using one command!

:star2: **Cool things incoming!** v2 will bring many changes and make the tool much more useful. It's being [developed](https://github.com/Tutrox/scss-powertools) right now.

## Installation
### As a development dependency

`npm install scss-powertools --save-dev`

### Use once (locally)

`npx scss-powertools <input> <output> [options]` No install needed!

_or_

`npm install scss-powertools --global` For older versions of NPM.

## Usage
**scss-powertools is made really simple**, and only consists of one command:

```
scss-powertools <input> <output> [options]
```

`input (SCSS)` and `output (CSS)` are references to your input SCSS and output CSS, relative to your project root (or where the command is run). If you have your SCSS in `scss/app.scss` and want to output to `dist/styles.css`, your command will look like:

```
scss-powertools scss/app.scss dist/styles.css
```

#### Options
Currently there are two options. These should **not** be combined.

```
-p or --production => Minify the output CSS,
                      disable source maps and ERROR (non-zero exit code)
                      if any issues, like linting issues (use on your CI)
-m or --minify     => Minify the output CSS, even though you are on dev enviroment
```

## Cool features

### SCSS imports can resolve to the `node_modules` folder

Did you write your imports like this earlier?

```
@import "node_modules/bootstrap";
```

No need to, anymore. Just write:

```
@import "bootstrap";
```

Easy!

### Use in your CI-enviroment

Running `scss-powertools` in your CI is easy. Just make sure to include the **`--production`** flag. It will make sure that your CI build will error if anything happens (like a lint issue).

### No config needed

`scss-powertools` does not need any config. Everything from linting to minifying is preconfigured using recommended settings. You can find them in [`powertools.js`](https://github.com/Tutrox/scss-powertools/blob/master/lib/powertools.js).


---

¹Only in production
