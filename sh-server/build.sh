#!/bin/bash

cd assets && ./build.sh && cd ..
go-bindata-assetfs -fs assets/images/*.svg assets/*.html assets/style/style.css assets/script/script.js assets/script/deps.js
