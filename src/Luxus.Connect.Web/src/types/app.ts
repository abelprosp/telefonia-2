import { z } from 'zod';

export const ThemeSchema = z.enum(['dark', 'light', 'system']);

export type Theme = z.infer<typeof ThemeSchema>;
