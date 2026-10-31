import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;

public class CalculateAverageTest {
    public static void main(String[] args) throws Exception {
        String[] values = {"-99.9", "-12.3", "-1.5", "-1.0", "0.0", "0.3", "12.3", "99.9"};
        short[] expected = {-999, -123, -15, -10, 0, 3, 123, 999};
        for (int i = 0; i < values.length; i++) {
            byte[] input = ("name;" + values[i] + "\n").getBytes(StandardCharsets.UTF_8);
            short actual = CalculateAverage.parseNumber(input, 5, values[i].length());
            if (actual != expected[i]) {
                throw new AssertionError(values[i] + ": expected " + expected[i] + ", actual " + actual);
            }
        }
        var file = Files.createTempFile("1brc-test-", ".txt");
        var output = new ByteArrayOutputStream();
        var original = System.out;
        try {
            Files.writeString(file, "Zurich;12.3\nAlpha;-1.0\nZurich;-2.3\nAlpha;3.0\n");
            try (var captured = new PrintStream(output, true, StandardCharsets.UTF_8)) {
                System.setOut(captured);
                CalculateAverage.main(new String[]{file.toString()});
            } finally {
                System.setOut(original);
            }
            String actual = output.toString(StandardCharsets.UTF_8).trim();
            String expectedOutput = "{Alpha=-1.0/1.0/3.0, Zurich=-2.3/5.0/12.3}";
            if (!actual.equals(expectedOutput)) {
                throw new AssertionError("expected " + expectedOutput + ", actual " + actual);
            }
        } finally {
            Files.deleteIfExists(file);
        }
        System.out.println("Java parser and aggregation tests passed");
    }
}
