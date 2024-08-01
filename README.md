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
