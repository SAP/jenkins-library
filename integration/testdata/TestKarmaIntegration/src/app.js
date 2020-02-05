#!/usr/bin/env node
'use strict';
var express = require('express');
var app = express();
var cfenv = require('cfenv')
var appEnv = cfenv.getAppEnv()

var home = require('./backend/home.js');
var getData = require('./backend/dataMgmt.js');
var getProcessingStatus = require('./backend/status.js');

app.get('/', home);

app.get('/data', getData);

app.get('/status', getProcessingStatus);

app.post('/', home);

app.use(express.static('src/frontend'));


module.exports = app.listen(appEnv.port, appEnv.bind, function () {
    console.log('server starting on ' + appEnv.url);
});;
