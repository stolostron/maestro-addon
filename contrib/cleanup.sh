#!/bin/bash

rm -f *.pem
rm -f *.csr
rm -f join.sh

kind delete clusters demo
