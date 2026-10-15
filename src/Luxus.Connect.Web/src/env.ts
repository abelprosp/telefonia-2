import { z } from 'zod';

const envSchema = z.object({
  VITE_API_URL: z.string(),
  VITE_STORAGE_BUCKET_NAME: z.string()
});

export const env = envSchema.parse(import.meta.env);
