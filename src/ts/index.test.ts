import { test, expect } from "bun:test";
import { mkdtempSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const entry = resolve(process.env.BRC_ENTRY ?? "index.ts");
test("CLI aggregates repeated stations, signs, extremes and sorts output through workers", () => {
    const dir = mkdtempSync(join(tmpdir(), "1brc-"));
    try {
        const file = join(dir, "measurements.txt");
        // Small chunks split records across reads; the final line has no newline.
        writeFileSync(file, "Zurich;12.3\nAlpha;-1.0\nZurich;-2.3\nAlpha;3.0\nExtreme;-99.9\nExtreme;99.9\n".repeat(20000).trimEnd());
        const result = spawnSync(process.execPath, [entry, file, "1"], { encoding: "utf8", timeout: 20000 });
        expect(result.error).toBeUndefined();
        expect(result.status).toBe(0);
        expect(result.stderr).toBe("");
        expect(result.stdout.trim()).toBe("Alpha=-1.0/1.0/3.0\nExtreme=-99.9/0.0/99.9\nZurich=-2.3/5.0/12.3");
    } finally { rmSync(dir, { recursive: true, force: true }); }
}, 25000);

test("CLI reports missing input files with a failing exit code", () => {
    const dir = mkdtempSync(join(tmpdir(), "1brc-missing-"));
    try {
        const result = spawnSync(process.execPath, [entry, join(dir, "missing.txt")], { encoding: "utf8", timeout: 20000 });
        expect(result.error).toBeUndefined();
        expect(result.status).not.toBe(0);
        expect(result.stderr).toContain("ENOENT");
    } finally { rmSync(dir, { recursive: true, force: true }); }
}, 25000);
