#!/bin/bash -e +x

go build -o ./build/brc ./src/go/*.go

# time ./build/brc -file /Users/mlee/Downloads/data/measurements_sm.txt
# time ./build/brc -file /Users/mlee/Downloads/data/accra.txt
# time ./build/brc -file /Users/mlee/Downloads/data/abasingammedda.txt
# time ./build/brc -file /Users/mlee/Downloads/data/measurements_med.txt
time ./build/brc -file /Users/mlee/Downloads/data/measurements.txt  #-cpuprofile large.prof