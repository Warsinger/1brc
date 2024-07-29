#!/bin/bash -e +x

javac -sourcepath src/java -d build/java src/java/CalculateAverage.java src/java/CalculateAverageTest.java

# time java -cp build/java CalculateAverage ./data/weather_stations.csv
# time java -Xmx6g -cp build/java CalculateAverage /Users/mlee/Downloads/data/measurements.txt
# time java -cp build/java CalculateAverage /Users/mlee/Downloads/data/measurements_a.txt
# time java -cp build/java CalculateAverage /Users/mlee/Downloads/data/measurements_sm.txt

# java -cp build/java CalculateAverageTest


jbang --javaagent=ap-loader@jvm-profiling-tools/ap-loader=start,event=cpu,file=profile.html src/java/CalculateAverage.java /Users/mlee/Downloads/data/measurements.txt