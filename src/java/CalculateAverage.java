/*
 *  Copyright 2023 The original authors
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;
import java.util.TreeMap;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Future;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collector;
import static java.util.stream.Collectors.groupingBy;

public class CalculateAverage {

    private static final byte EOL = 10;
    private static final byte SEMICOLON = 59;
    private static final byte DASH = 45;
    private static final byte DOT = 46;
    private static final byte ZERO = 48;
    private static final int READ_SIZE = 8192 * 8192;

    private static Map<String, Result> processChunk(byte[] chunk, int endOfChunk) {
        if (chunk[0] == SEMICOLON) {
            var message = String.format("no name at start of chunk: {%s}\n", new String(chunk, 0, chunk.length - (chunk.length - endOfChunk)));
            System.out.println(message);
            throw new RuntimeException(message);
        } else if (chunk[endOfChunk] != EOL) {
            var message = String.format("no eol at end of chunk %d: {%d}\n", endOfChunk, chunk[endOfChunk]);
            System.out.println(message);
            throw new RuntimeException(message);
        }
        var measurements = new HashMap<String, Result>(100);
        int lineIdx = 0;
        while (lineIdx < endOfChunk) {
            // read till EOL
            int nameStart = lineIdx;
            int nameIdx = lineIdx;
            for (; nameIdx < endOfChunk && chunk[nameIdx] != SEMICOLON; ++nameIdx) {
                // first part is the name
            }
            String station = new String(chunk, nameStart, nameIdx - nameStart);
            int tempStart = nameIdx + 1;
            int tempIdx = tempStart;
            for (; tempIdx <= endOfChunk && chunk[tempIdx] != EOL; ++tempIdx) {
                // second part is the temp
            }
            short temp = parseNumber(chunk, tempStart, tempIdx - tempStart);
            lineIdx = tempIdx + 1;

            var result = measurements.get(station);
            if (result == null) {
                // System.out.printf("new station found %s;%d\bytesRead", station, temp);
                measurements.put(station, new Result(temp));
                // System.exit(1);
            } else {
                // System.out.printf("existing station found %s;%d;%s\bytesRead", station, temp, result);
                result.accumulate(temp);
                // System.out.printf("new result %s\bytesRead", result);
                // System.exit(1);
            }
        }
        // System.out.println("done with measurements");
        return measurements;
    }

    static short parseNumber(byte[] chunk, int tempIdx, int len) {
        int startIndex = tempIdx;
        // System.out.printf("number %s, idx %d, len %d\bytesRead", new String(chunk, tempIdx, len), tempIdx, len);
        boolean isNegative = false;
        if (chunk[tempIdx] == DASH) {
            isNegative = true;
            ++tempIdx;
            // System.out.printf("idx %d\bytesRead", tempIdx);
        }
        // System.out.println("negative " + isNegative);
        short sum;
        if (chunk[tempIdx + 1] == DOT) {
            // single digit number
            sum = (short) ((chunk[tempIdx] - ZERO) * 10);
            tempIdx += 2;
            // System.out.printf("idx %d sum %d\bytesRead", tempIdx, sum);
        } else {
            // 2 digit number
            sum = (short) ((chunk[tempIdx] - ZERO) * 100 + (chunk[tempIdx + 1] - ZERO) * 10);
            // System.out.printf("idx %d sum %d\bytesRead", tempIdx, sum);
            tempIdx += 3;
        }
        if (tempIdx >= READ_SIZE) {
            System.out.printf("startidx %d idx %d, chunk %s\n", startIndex, tempIdx, new String(chunk, 0, tempIdx - 1));
        }
        sum += chunk[tempIdx] - ZERO;
        // System.out.printf("idx %d sum %d, d %d\bytesRead", tempIdx, sum, chunk[tempIdx]);

        if (isNegative) {
            sum = (short) -sum;
        }

        return sum;
    }

    private static byte[] recycle(byte[] chunk, int eol, int len) {
        // return a new byte array of same size as chunk and copy the end of chunk into the first of new chunk
        // System.out.printf("recycling chunk: %d, %d\bytesRead", eol, len);
        // System.out.printf("recycling chunk: {%s}\bytesRead", new String(chunk, eol, len));
        var newChunk = new byte[chunk.length];
        System.arraycopy(chunk, eol, newChunk, 0, len);
        // System.out.printf("recycled chunk: {%s}\bytesRead", new String(newChunk, 0, len));
        return newChunk;
    }

    public static ExecutorService newBlockingThreadPool(int threadCount) {
        return new ThreadPoolExecutor(threadCount, threadCount,
                0L, TimeUnit.MILLISECONDS,
                new LinkedBlockingQueue<>(threadCount),
                new ThreadPoolExecutor.CallerRunsPolicy());
    }

    private static Map<String, Result> calculate(String fileName) throws IOException, InterruptedException, ExecutionException {
        try (var input = Files.newInputStream(Paths.get(fileName), StandardOpenOption.READ)) {
            var measurements = new TreeMap<String, Result>();
            var chunk = new byte[READ_SIZE];
            var remaining = 0;
            var cores = Runtime.getRuntime().availableProcessors() + 1;
            var exs = newBlockingThreadPool(cores);
            var futures = new ArrayList<Future<Map<String, Result>>>(1000);
            while (true) {
                var bytesRead = input.read(chunk, remaining, chunk.length - remaining);
                var amountInBuf = bytesRead + remaining;
                if (amountInBuf > 0) {
                    // find ending \n in the chunk
                    final var endOfChunk = findLastEOL(chunk, amountInBuf);
                    var chunk2 = chunk;
                    // send chunk for processing
                    futures.add(exs.submit(() -> processChunk(chunk2, endOfChunk)));
                    // move past eol
                    var newChunkStart = endOfChunk + 1;
                    // how much data is left at the end
                    remaining = amountInBuf - newChunkStart;
                    // System.out.printf("amount %d remaining %d\bytesRead", amountInBuf, remaining);
                    if (remaining > 0) {
                        // copy end to begining for next iteration
                        chunk = recycle(chunk, newChunkStart, remaining);
                    } else {
                        remaining = 0;
                        chunk = new byte[READ_SIZE];
                    }
                } else {
                    break;
                }
            }
            // System.err.println("futures " + futures.size());
            for (var f : futures) {
                var results = f.get();
                aggregateResults(measurements, results);
            }
            exs.shutdown();
            return measurements;
        }
    }

    private static void aggregateResults(Map<String, Result> measurements, Map<String, Result> results) {
        for (var e : results.entrySet()) {
            String key = e.getKey();
            var existing = measurements.get(key);
            if (existing == null) {
                measurements.put(key, e.getValue());
            } else {
                var value = e.getValue();
                existing.accumulate(value);
            }
        }
    }

    private static int findLastEOL(byte[] b, int n) {
        for (int i = n - 1; i >= 0; --i) {
            if (b[i] == EOL) {
                return i;
            }
        }
        throw new RuntimeException("we should have found an EOL");
    }

    public static void main(String[] args) throws IOException, InterruptedException, ExecutionException {
        if (args.length < 1) {
            throw new RuntimeException("file name not specified");
        }
        var fileName = args[0];

        var measurements = calculate(fileName);
        // var measurements = calculateInitial(fileName);

        System.out.println(measurements);
    }

    private static record Measurement(String station, double value) {

        private Measurement(String[] parts) {
            this(parts[0], Double.parseDouble(parts[1]));
        }
    }

    private static final class Result {

        short min;
        short max;
        int count;
        long sum;

        Result(short temp) {
            this.min = temp;
            this.max = temp;
            this.count = 1;
            this.sum = temp;
            // System.out.printf("new Result %d, %s\n", temp, Result.this);
        }

        void accumulate(short temp) {
            if (temp < min) {
                min = temp;
            } else if (temp > max) {
                max = temp;
            }
            sum += temp;
            ++count;
            // System.out.printf("acc temp Result %d, %s\n", temp, Result.this);
        }

        void accumulate(Result other) {
            if (other.min < min) {
                min = other.min;
            }
            if (other.max > max) {
                max = other.max;
            }
            sum += other.sum;
            count += other.count;
            // System.out.printf("acc other Result %s, %s\n", other, Result.this);
        }

        @Override
        public String toString() {
            return round10(min) + "/" + round(sum / 10.0 / count) + "/" + round10(max);// + "{" + sum + "/" + count + "}";
        }
    }

    private static record ResultRow(double min, double mean, double max) {

        @Override
        public String toString() {
            return round(min) + "/" + round(mean) + "/" + round(max);
        }

    }

    private static double round(double value) {
        return Math.round(value * 10.0) / 10.0;
    }

    private static double round10(double value) {
        return Math.round(value) / 10.0;
    }

    private static class MeasurementAggregator {

        private double min = Double.POSITIVE_INFINITY;
        private double max = Double.NEGATIVE_INFINITY;
        private double sum;
        private long count;
    }

    private static Map<String, ResultRow> calculateInitial(String fileName) throws IOException {
        Collector<Measurement, MeasurementAggregator, ResultRow> collector = Collector.of(
                MeasurementAggregator::new,
                (a, m) -> {
                    a.min = Math.min(a.min, m.value);
                    a.max = Math.max(a.max, m.value);
                    a.sum += m.value;
                    a.count++;
                },
                (agg1, agg2) -> {
                    var res = new MeasurementAggregator();
                    res.min = Math.min(agg1.min, agg2.min);
                    res.max = Math.max(agg1.max, agg2.max);
                    res.sum = agg1.sum + agg2.sum;
                    res.count = agg1.count + agg2.count;

                    return res;
                },
                agg -> {
                    return new ResultRow(agg.min, (Math.round(agg.sum * 10.0) / 10.0) / agg.count, agg.max);
                });
        Map<String, ResultRow> measurements = new TreeMap<>(Files.lines(Paths.get(fileName))
                .map(l -> new Measurement(l.split(";")))
                .collect(groupingBy(m -> m.station(), collector)));
        return measurements;
    }
}
