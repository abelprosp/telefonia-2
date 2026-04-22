import { defineConfig } from '@cerios/openapi-to-zod';

export default defineConfig({
  defaults: {
    mode: 'strict',
    includeDescriptions: true,
    useDescribe: false,
    showStats: true,
    schemaType: 'all'
  },
  specs: [
    {
      input: './docs/openapi.yaml',
      outputTypes: 'src/api/schemas.ts'
    }
  ]
});
