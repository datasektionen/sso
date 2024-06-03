import babel from "@rollup/plugin-babel";
import resolve from "@rollup/plugin-node-resolve";
import typescript from "@rollup/plugin-typescript";
import { readdirSync } from "node:fs";

/**
 * @type {import("rollup").RollupOptions}
 */
export default {
    input: readdirSync("islands")
        .filter(filename => filename.endsWith(".island.tsx"))
        .map(filename => `islands/${filename}`),
    output: {
        dir: "dist",
        format: "es",
    },
    plugins: [
        typescript(),
        resolve(),
        babel({
            babelHelpers: "bundled",
            extensions: ["ts", "tsx"],
            presets: ["solid"],
        }),
    ],
};
