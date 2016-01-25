@echo off

pushd processing

@echo on

python detectBlur.py ../%1

@echo off

popd

@echo on
