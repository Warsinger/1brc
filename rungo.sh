#!/bin/bash -e +x

go build -o ./build/brc ./src/go/*.go

# time ./build/brc /Users/mlee/Downloads/data/measurements_sm.txt
# time ./build/brc /Users/mlee/Downloads/data/accra.txt
# time ./build/brc /Users/mlee/Downloads/data/abasingammedda.txt
# time ./build/brc /Users/mlee/Downloads/data/measurements_med.txt
time ./build/brc /Users/mlee/Downloads/data/measurements.txt