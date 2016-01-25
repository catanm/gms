#!/bin/bash

pushd processing > /dev/null

if [ ! -a ImagePHash.class ]; then
	javac ImagePHash.java
fi

java ImagePHash ../$1

popd > /dev/null
