/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        devman: {
          bg: '#0F172A',
          'bg-soft': '#111B2E',
          'bg-deep': '#08111E',
          panel: '#101A2B',
          'panel-raised': '#16243B',
          'panel-muted': '#1A2A43',
          border: '#233552',
          'border-strong': '#35507A',
          accent: '#22C55E',
          info: '#38BDF8',
          warning: '#F59E0B',
          danger: '#EF4444',
          'text-primary': '#F8FAFC',
          'text-muted': '#94A3B8',
        },
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'SF Mono', 'Consolas', 'monospace'],
        sans: ['IBM Plex Sans', 'Inter', 'SF Pro Text', 'Segoe UI', 'Noto Sans', 'Arial', 'sans-serif'],
      },
      borderRadius: {
        'card': '22px',
        'panel': '24px',
        'badge': '12px',
        'button': '12px',
        'input': '16px',
      },
    },
  },
  plugins: [],
}
