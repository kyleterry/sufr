var chalk = require("chalk");
var successMsg = chalk.green;
var errorMsg = chalk.red;
var importantMsg = chalk.white.inverse;
var boldMsg = chalk.bold;
var _ = require("lodash");

var config;

function getConfig(data){
  return config[data];
}

function setConfig(data){
  config = data;
}

/* eslint-disable no-console */

function log(msg){
  console.log(msg);
}

function success(msg, service){
  console.log(`${successMsg("Success in step:")} ${importantMsg(service)}`);
  console.log(msg);
}

function error(msg, service){
  console.error(`${errorMsg("Error in step:")} ${importantMsg(service)}`);
  if (service == "lint"){
    console.log(`In file ${boldMsg(config.input)}`);
    _.forEach(JSON.parse(msg)[0].warnings, value => {
      console.log(`${value.text} on line ${value.line}`);
    });
  } else{
    console.error(msg);
  }
  if (config.production) process.exitCode = 1;
}

module.exports = {
  getConfig,
  setConfig,
  log,
  success,
  error
};
