import * as fs from 'fs';
import * as readline from 'readline';

interface TemperatureData {
    min: number;
    max: number;
    sum: number;
    count: number;
}

const aggregateTemperatures = async (filePath: string) => {
    const fileStream = fs.createReadStream(filePath);
    const rl = readline.createInterface({
        input: fileStream,
        crlfDelay: Infinity
    });

    const temperatureMap: Record<string, TemperatureData> = {};

    for await (const line of rl) {
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

    const names = Object.keys(temperatureMap).sort();

    for (const name of names) {
        const data = temperatureMap[name];
        const avg = data.sum / data.count;
        console.log(`${name}: min=${data.min.toFixed(1)}, max=${data.max.toFixed(1)}, avg=${avg.toFixed(1)}`);
    }
};

const main = () => {
    const args = process.argv.slice(2);
    if (args.length !== 1) {
        console.error('Usage: bun run ts-node index.ts <file-path>');
        process.exit(1);
    }

    const filePath = args[0];
    aggregateTemperatures(filePath).catch(err => console.error(err));
};

main();