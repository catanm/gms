#!/bin/bash

pushd processing > /dev/null

python detectBlur.py ../$1

popd > /dev/null
