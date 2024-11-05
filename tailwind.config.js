/** @type {import("tailwindcss").Config} */
export default {
    content: ["./templates/**/*.templ"],
    theme: {
        extend: {
            colors: {
                cerise: {
                    strong: "#ee2a7b",
                    regular: "#e83d84",
                    light: "#ec5f99",
                },
            }
        },
    },
    plugins: [],
}
