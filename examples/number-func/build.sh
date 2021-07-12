#!/bin/bash
cat ./number-func.go | ../../genny gen "number=float32,int32" > number-func-gen.go
