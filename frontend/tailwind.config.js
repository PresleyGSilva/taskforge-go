/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts}"],
  theme: {
    extend: {
      colors: { brand: { DEFAULT: "#7C3AED", dark: "#6D28D9", light: "#A78BFA" } },
    },
  },
  plugins: [],
};
