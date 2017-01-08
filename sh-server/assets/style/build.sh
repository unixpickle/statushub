#!/bin/bash

if [ ! -d node_modules ]; then
  npm install
fi

./node_modules/less/bin/lessc src/index.less style.css
