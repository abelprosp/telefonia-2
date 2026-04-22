import { defineConfig } from '@kubb/core';
import { pluginClient } from '@kubb/plugin-client';
import { pluginOas } from '@kubb/plugin-oas';
import { pluginReactQuery } from '@kubb/plugin-react-query';
import { pluginTs } from '@kubb/plugin-ts';

export default defineConfig({
  root: '.',
  input: {
    path: './docs/openapi.yaml'
  },
  output: {
    path: './src/api',
    clean: true
  },
  plugins: [
    pluginOas(),
    pluginTs(),
    pluginReactQuery(),
    pluginClient({
      importPath: '@/lib/client.ts'
    })
  ]
});
