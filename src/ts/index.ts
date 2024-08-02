import * as fs from 'fs';
import * as os from 'os';
import { Worker, isMainThread, parentPort, workerData } from 'worker_threads';

interface TemperatureData {
    min: number;
    max: number;
    sum: number;
    count: number;
}

const CHUNK_SIZE = 32 * 1024 * 1024; // 32MB default chunk size
const NUM_CORES = os.cpus().length;

if (isMainThread) {
    const processChunk = (chunk: string): Record<string, TemperatureData> => {
        const temperatureMap: Record<string, TemperatureData> = {};
        const lines = chunk.split('\n');
        for (const line of lines) {
            if (!line.trim()) continue; // Skip empty lines
            const [name, tempStr] = line.split(';');
            const temp = parseFloat(tempStr);

            if (!temperatureMap[name]) {
                temperatureMap[name] = { min: temp, max: temp, sum: temp, count: 1 };
            } else {
                const data = temperatureMap[name];
                data.min = Math.min(data.min, temp);
                data.max = Math.max(data.max, temp);
                data.sum += temp;
                data.count += 1;
            }
        }
        return temperatureMap;
    };

    const mergeTemperatureMaps = (map1: Record<string, TemperatureData>, map2: Record<string, TemperatureData>): Record<string, TemperatureData> => {
        for (const name in map2) {
            if (map1[name]) {
                map1[name].min = Math.min(map1[name].min, map2[name].min);
                map1[name].max = Math.max(map1[name].max, map2[name].max);
                map1[name].sum += map2[name].sum;
                map1[name].count += map2[name].count;
            } else {
                map1[name] = map2[name];
            }
        }
        return map1;
    };

    const aggregateTemperatures = async (filePath: string, chunkSize: number) => {
        const temperatureMap: Record<string, TemperatureData> = {};
        const fileStream = fs.createReadStream(filePath, { highWaterMark: chunkSize });

        let buffer = '';
        let workers: Worker[] = [];

        fileStream.on('data', (chunk) => {
            buffer += chunk.toString();

            let lastNewLineIndex = buffer.lastIndexOf('\n');
            if (lastNewLineIndex !== -1) {
                let processableChunk = buffer.substring(0, lastNewLineIndex);
                buffer = buffer.substring(lastNewLineIndex + 1);

                if (workers.length < NUM_CORES) {
                    const worker = new Worker(__filename, {
                        workerData: { chunk: processableChunk }
                    });

                    worker.on('message', (message) => {
                        mergeTemperatureMaps(temperatureMap, message);
                        workers = workers.filter(w => w !== worker);
                    });

                    workers.push(worker);
                } else {
                    processChunk(processableChunk);
                }
            }
        });

        fileStream.on('end', async () => {
            if (buffer.length > 0) {
                processChunk(buffer);
            }

            await Promise.all(workers.map(worker => new Promise(resolve => worker.on('exit', resolve))));

            const sortedNames = Object.keys(temperatureMap).sort();

            for (const name of sortedNames) {
                const data = temperatureMap[name];
                const avg = data.sum / data.count;
                console.log(`${name}=${data.min.toFixed(1)}/${avg.toFixed(1)}/${data.max.toFixed(1)}`);
            }
        });

        fileStream.on('error', (err) => {
            console.error('Error reading file:', err);
        });
    };

    const main = () => {
        const args = process.argv.slice(2);
        if (args.length < 1 || args.length > 2) {
            console.error('Usage: bun run ts-node index.ts <file-path> [chunk-size-in-mb]');
            process.exit(1);
        }

        const filePath = args[0];
        const chunkSize = args.length === 2 ? parseInt(args[1], 10) * 1024 * 1024 : CHUNK_SIZE;
        aggregateTemperatures(filePath, chunkSize).catch(err => console.error(err));
    };

    main();
} else {
    const processChunk = (chunk: string): Record<string, TemperatureData> => {
        const temperatureMap: Record<string, TemperatureData> = {};
        const lines = chunk.split('\n');
        for (const line of lines) {
            if (!line.trim()) continue;
            const [name, tempStr] = line.split(';');
            const temp = parseFloat(tempStr);

            if (!temperatureMap[name]) {
                temperatureMap[name] = { min: temp, max: temp, sum: temp, count: 1 };
            } else {
                const data = temperatureMap[name];
                data.min = Math.min(data.min, temp);
                data.max = Math.max(data.max, temp);
                data.sum += temp;
                data.count += 1;
            }
        }
        return temperatureMap;
    };

    const temperatureMap = processChunk(workerData.chunk);
    parentPort.postMessage(temperatureMap);
}