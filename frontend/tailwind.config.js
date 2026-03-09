/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",  // 让插件和编译器能扫描这些文件
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
};

