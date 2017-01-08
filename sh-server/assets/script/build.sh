#!/bin/bash

if [ ! -d node_modules ]; then
  npm install
fi

if [ ! -d deps ]; then
  mkdir deps
  curl https://unpkg.com/react@15.4.2/dist/react.js >deps/react.js
  curl https://unpkg.com/react-dom@15.4.2/dist/react-dom.js >deps/react-dom.js
fi

cat src/root.js >joined.js
cat src/client.js >>joined.js
cat src/loader.js >>joined.js
cat src/overview.js >>joined.js
cat src/nav_bar.js >>joined.js

cat deps/react.js >deps.js
cat deps/react-dom.js >>deps.js

node ./node_modules/babel-cli/bin/babel.js joined.js --plugins \
  transform-react-jsx --out-file script.js && rm joined.js
