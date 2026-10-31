# 1brc

[1 Billion Row Challenge](https://github.com/gunnarmorling/1brc/blob/main/README.md)

Running the 1brc in Java and Go and ???

## Java

Only reading the file with raw input stream reads of 64MB is ~3s

Timings are on Apple MBPro M1 16GB RAM

| Change | Timing |
| --- | --- |
| Initial basline on my local machine (battery) | 3m 59s |
| Initial single threaded straight read and process | 1m 22s |
| Read and hand off to thread for processing + 6G heap | 30s |
| 16MB read size | 18s |
| 64MB read size | 15s |
| ~~128MB read size~~ | 15s |
| Reduce Result allocation | 13s |

## Go

Use Go 1.27.3 or newer, matching `go.mod`. Build the Go executable from the
repository root with `go build -o ./build/brc ./src/go`.

Reading only using Scanner with defaults, took ~23s.
Reading 64MB chunks at a time is ~3s

| Change | Timing |
| --- | --- |
| Initial basline, singlethread  | 2m 57s |
| First pass threadding and channelsl; read size 64mb| 10.25 |
| Read size 32mb | 9.8s |
| ~~chunks yield arrays rather than maps~~ | 1m 9s |
| ~~read via memory mapping ~~ | 29s |
| ~~read results into arrays rather than channels ~~ | 10.8s |
| process chunk using custom hashing from [official repo go submission](https://github.com/gunnarmorling/1brc/blob/4daeb94b048e074c2b80aac1074b68eb92285ea8/src/main/go/AlexanderYastrebov/calc.go#L132) | 5.1s |
| don't check for hash collisions | 4.6s |
| precalculate constants | 4.4s |

## TypeScript

I had ChatGPT generate the baseline code with this prompt
>> i want a basic typescript project that reads in a set of lines from a file that contain a name and a number separated by a semicolon. the numbers are temperatures ranging from -99.9 to 99.9 and have only one decimal point
the results should be aggregated by name to find the min, max, and average number

Then I told it I wanted to use `bun` and needed to pass in the file name.

The code worked first pass. Just need to adjust the formatting.
When run on the large dataset, it used up all the memory. 2x of my RAM. After a couple of prompts, chatty gave me some code that ran in pretty much constant memory ~86MB.

| Change | Timing |
| --- | --- |
| Baseline ChatGPT, one line at a time | 4m 52s |
| Read file by lines but process in chunks of 32MB | 5m 47s|
| Read file in chunks of 32M and process | 4m 24s |
| First multithreaded attempt | 5m 6s |
| Second multithreaded attempt | 4m 24s |

## Build and test

CI runs separate Go, Java and TypeScript jobs on pull requests, pushes to main,
and manual dispatch. Standard Actions use major version tags. Tests use small
fixtures; they do not run the billion-row benchmark or enforce performance timings.

Go (version from `go.mod`):

```sh
go test -race ./...
go vet ./...
go mod verify
go mod tidy -diff
go build -o build/brc ./src/go
```

The parser regression exercises `processChunk`, including signed one- and
two-digit temperatures, rather than the removed `parseTemp` helper.
The race detector requires a supported C toolchain.

Java (JDK 27 or newer; CI uses the latest Temurin 27 patch):

```sh
mkdir -p build/java
javac --release 27 -d build/java src/java/*.java
java -cp build/java CalculateAverageTest
```

The Java tests throw on parser/aggregation failures so CI exits unsuccessfully.

TypeScript 7.0.2 (Bun 1.4.4; locked development dependencies):

```sh
cd src/ts
bun install --frozen-lockfile
bunx tsc --noEmit
bun test
mkdir -p ../../build/ts
bun build index.ts --target=bun --outfile=../../build/ts/index.js
BRC_ENTRY=../../build/ts/index.js bun test
```

TypeScript tests run the real CLI and workers against both source and built
JavaScript. They cover repeated stations, sorted output, signed/extreme values,
records split across reads, a final line without a newline, and missing input
files. Input/worker errors now produce a nonzero CLI exit code.

The historical run scripts still contain the author's local
benchmark input paths; edit those paths before running a benchmark.
