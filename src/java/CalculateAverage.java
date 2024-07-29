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
import java.util.HashMap;
import java.util.Map;
import java.util.TreeMap;
import java.util.stream.Collector;
import static java.util.stream.Collectors.groupingBy;

public class CalculateAverage {

    private static final byte EOL = 10;
    private static final byte SEMICOLON = 59;
    private static final byte DASH = 45;
    private static final byte DOT = 46;
    private static final byte ZERO = 48;
    private static final int READ_SIZE = 1024 * 1024;

    private static void processChunk(byte[] chunk, int endOfChunk, Map<String, Result> measurements) {
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
            for (; tempIdx < endOfChunk && chunk[tempIdx] != EOL; ++tempIdx) {
                // second part is the temp
            }
            int temp = parseNumber(chunk, tempStart, tempIdx - tempStart);
            lineIdx = tempIdx + 1;

            var result = measurements.get(station);
            if (result == null) {
                result = new Result(temp, temp, temp, 1);
                // System.out.printf("new station found %s;%d\n", station, temp);
                measurements.put(station, result);
                // System.exit(1);
            } else {
                int min = Math.min(result.min, temp);
                int max = Math.max(result.max, temp);
                int sum = result.sum + temp;
                int count = result.count + 1;
                measurements.put(station, new Result(min, max, sum, count));
            }
        }
    }

    private static int parseNumber(byte[] chunk, int tempIdx, int len) {
        // System.out.printf("number %s, idx %d, len %d\n", new String(chunk, tempIdx, len), tempIdx, len);
        boolean isNegative = false;
        if (chunk[tempIdx] == DASH) {
            isNegative = true;
            ++tempIdx;
            // System.out.printf("idx %d\n", tempIdx);
        }
        // System.out.println("negative " + isNegative);
        int sum;
        if (chunk[tempIdx + 1] == DOT) {
            // single digit number
            sum = chunk[tempIdx] - ZERO;
            tempIdx += 2;
        } else {
            // 2 digit number
            sum = (chunk[tempIdx] - ZERO) * 100 + (chunk[tempIdx + 1] - ZERO) * 10;
            // System.out.printf("idx %d sum %d\n", tempIdx, sum);
            tempIdx += 3;
        }
        sum += chunk[tempIdx] - ZERO;
        // System.out.printf("idx %d sum %d, d %d\n", tempIdx, sum, chunk[tempIdx]);

        if (isNegative) {
            sum = -sum;
        }

        return sum;
    }

    private static byte[] recycle(byte[] chunk, int eol, int len) {
        // Try skipping the copy until we parallelize and need to make a safety copy
        // return a new byte array of same size as chunk and copy the end of chunk into the first of new chunk
        var newChunk = chunk; // new byte[chunk.length];
        System.arraycopy(chunk, 0, newChunk, eol, len - 1);
        return newChunk;
    }

    private static Map<String, Result> readFile(String fileName) throws IOException {
        try (var input = Files.newInputStream(Paths.get(fileName), StandardOpenOption.READ)) {
            var measurments = new HashMap<String, Result>(10000);
            var chunk = new byte[READ_SIZE];
            int offset = 0;
            while (true) {
                int n = input.read(chunk, offset, chunk.length - offset);
                var amountInBuf = n + offset;
                if (amountInBuf > 0) {
                    // processChunk data

                    // find ending \n in the chunk
                    int endOfChunk = findLastEOL(chunk, amountInBuf);

                    // send chunk for processing
                    processChunk(chunk, endOfChunk, measurments);
                    // how much data is left at the end
                    offset = amountInBuf - endOfChunk;
                    // copy end to begining for next iteration
                    chunk = recycle(chunk, endOfChunk + 1, offset);
                } else {
                    break;
                }
            }
            return measurments;
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

    public static void main(String[] args) throws IOException {
        if (args.length < 1) {
            throw new RuntimeException("file name not specified");
        }
        var fileName = args[0];

        var measurements = readFile(fileName);
        // var measurements = calculate(fileName);

        System.out.println(measurements);
    }

    private static record Measurement(String station, double value) {

        private Measurement(String[] parts) {
            this(parts[0], Double.parseDouble(parts[1]));
        }
    }

    private static record Result(int min, int max, int sum, int count) {
        // TODO implment formatting /10 and round to 1 digit
        // @Override
        // public String toString() {
        //     return round(min) + "/" + round(mean) + "/" + round(max);
        // }

    }

    private static record ResultRow(double min, double mean, double max) {

        @Override
        public String toString() {
            return round(min) + "/" + round(mean) + "/" + round(max);
        }

        private double round(double value) {
            return Math.round(value * 10.0) / 10.0;
        }
    }

    ;

    private static class MeasurementAggregator {

        private double min = Double.POSITIVE_INFINITY;
        private double max = Double.NEGATIVE_INFINITY;
        private double sum;
        private long count;
    }

    private static Map<String, ResultRow> calculate(String fileName) throws IOException {
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
