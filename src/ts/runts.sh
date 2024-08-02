#!/bin/bash -e +x

bun  build ./index.ts --outdir ./build --target bun

# time bun run ./build/index.js /Users/mlee/Downloads/data/measurements.txt 32
time bun  ./build/index.js /Users/mlee/Downloads/data/measurements.txt 32