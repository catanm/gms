@echo off

pushd processing

if [ ! -a ImagePHash.class ]; then
	javac ImagePHash.java
fi

@echo on

java ImagePHash ../%1

@echo off

popd > /dev/null

@echo on
