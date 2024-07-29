#!/bin/bash -e +x

javac -sourcepath src/java -d build/java src/java/CalculateAverage.java

time java -cp build/java CalculateAverage ./data/weather_stations.csv
# time java -cp build/java CalculateAverage /Users/mlee/Downloads/data/measurements.txt
