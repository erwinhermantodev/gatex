/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        background: '#0a0a0c',
        sidebar: '#0d0d0f',
        card: 'rgba(255, 255, 255, 0.03)',
        border: 'rgba(255, 255, 255, 0.08)',
        primary: {
          DEFAULT: '#06b6d4', // Cyan 500
          glow: 'rgba(6, 182, 212, 0.2)'
        },
        purple: {
          DEFAULT: '#a855f7',
          glow: 'rgba(168, 85, 247, 0.2)'
        },
        emerald: {
          DEFAULT: '#10b981',
          glow: 'rgba(16, 185, 129, 0.2)'
        }
      },
      backgroundImage: {
        'glass-gradient': 'linear-gradient(135deg, rgba(255, 255, 255, 0.05) 0%, rgba(255, 255, 255, 0.01) 100%)',
      }
    },
  },
  plugins: [],
}
