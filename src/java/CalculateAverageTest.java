
import static java.lang.System.out;

public class CalculateAverageTest {

    public static void main(String[] args) {
        byte[] b = {50, 46, 53, 10, 69, 67, 75};
        // out.println(new String(b));
        var expected = 25;
        var n = CalculateAverage.parseNumber(b, 0, 3);
        if (n != expected) {
            out.printf("expected %d, actual %d\n", expected, n);
        } else {
            out.println("success");
        }
    }
}
