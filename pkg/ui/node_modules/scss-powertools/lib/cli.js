#!/usr/bin/env node

var program = require("commander");
var chalk = require("chalk");
var _ = require("lodash");
var powerTools = require("./powertools");
var logger = require("./logger");

require("pkginfo")(module, "version");

var config = {
  production: false,
  minify: false,
  input: "",
  output: ""
};

program
  .usage("<source SCSS> <output CSS> [options]")
  .version(module.exports.version)
  .option("-p, --production", "Run in production, any error (such as lint) will fail the build. Also enables minify.")
  .option("-m, --minify", "Minify the file, even if not in production.")
  .parse(process.argv);

if (program.production) config.production = true;
if (program.minify) config.minify = true;

if (_.size(program.args) == 0){
  /* eslint-disable no-console */
  console.log(chalk.bold.yellow("scss-powertools!"));
  console.log(chalk.cyan("Usage:") + " scss-powertools --help");
  /* eslint-enable no-console */
} else if (_.size(program.args) != 2){
  logger.error("You have to give two arguments: <source> <output>", "initialize");
} else if (!(_.endsWith(program.args[0], ".scss")) || !(_.endsWith(program.args[1], ".css"))){
  logger.error("Some of your files do not have the right extension. Input should be .scss and output should be .css", "initialize");
} else{
  config.input = program.args[0];
  config.output = program.args[1];
  logger.setConfig(config);
  powerTools(config.input, config.output);
}
