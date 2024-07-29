# 1brc

1 Billion Row Challenge

Running the 1brc in Java and Go and ???

## Java

Only reading the file with raw input stream reads of 64MB is ~3s

There is some bug in the rounding that has things off by a small amount (~0.1-0.2 in some cases)

| Change | Timing |
| --- | --- |
| Initial basline on my local machine (battery) | 3m 59s |
| Initial single threaded straight read and process | 1m 22s |
| Read and hand off to thread for processing + 6G heap | 30s |
| 16MB read size | 18s |
| 64MB read size | 15s |
| ~~128MB read size~~ | 15s |
| Reduce Result allocation | 13s |
| General bug fixing ??? | 11s |

## Go

| Change | Timing |
| --- | --- |
| Initial basline  | ??? |
|  |  |
