import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  define: { 'process.env.baseUrl': JSON.stringify('') },
  server: {
    proxy: Object.fromEntries(['/leagues', '/teams', '/updates', '/source-data', '/dpc/standings', '/robots.txt', '/sitemap.xml', '/sitemaps/', '^/[0-9]+/'].map(path => [path, 'http://localhost:1323']))
  },
  build: {
    rollupOptions: {
      output: { entryFileNames: 'build/bundle.js', chunkFileNames: 'build/[name]-[hash].js', assetFileNames: 'build/[name][extname]' }
    }
  }
});
