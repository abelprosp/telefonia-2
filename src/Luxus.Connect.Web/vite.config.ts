import path from 'path';

import tailwindcss from '@tailwindcss/vite';
import { tanstackRouter } from '@tanstack/router-plugin/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// https://vite.dev/config/
export default defineConfig({
  optimizeDeps: {
    include: ['jwt-decode', 'react-oidc-context', 'oidc-client-ts']
  },
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    watch: {
      // Docker Desktop (Windows/macOS): file events às vezes não atravessam o bind mount
      usePolling: process.env.VITE_USE_POLLING === 'true'
    }
  },
  plugins: [
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
      generatedRouteTree: './src/route-tree.gen.ts',
      routesDirectory: './src/pages',
      routeToken: 'layout'
    }),
    react(),
    tailwindcss()
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  }
});
