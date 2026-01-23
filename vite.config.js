import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    outDir: 'static/dist',
    emptyOutDir: true,
    lib: {
      entry: 'assets/js/main.js',
      name: 'App',
      fileName: 'bundle',
      formats: ['iife'], // output as a single IIFE for script tag
    },
    rollupOptions: {
      output: {
        entryFileNames: 'bundle.js',
        extend: true, // Allow extending the global object (window)
      }
    },
    minify: true,
  },
});
