import { type JwtPayload } from 'jwt-decode';
import { z } from 'zod';

export const UserInfoSchema = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string(),
  avatar: z.string().optional(),
  timezone: z.string().optional(),
  roles: z.array(z.string()),
  organization: z.object({
    alias: z.string(),
    name: z.string()
  })
});

export const TokenInfoSchema = z.object({
  accessToken: z.string(),
  expiresIn: z.number(),
  refreshToken: z.string(),
  refreshExpiresIn: z.number()
});

export const TokenResponseSchema = z.object({
  access_token: z.string(),
  expires_in: z.number(),
  refresh_expires_in: z.number(),
  refresh_token: z.string(),
  token_type: z.string(),
  'not-before-policy': z.number(),
  session_state: z.string(),
  scope: z.string()
});

export type TokenPayload = JwtPayload & {
  name: string;
  email: string;
  family_name: string;
  given_name: string;
  roles?: string[];
  realm_access?: { roles: string[] };
  organization: Record<string, { name: string[] }>;
  resource_access: Record<string, { roles: string[] }>;
};

export type UserInfo = z.infer<typeof UserInfoSchema>;
export type TokenInfo = z.infer<typeof TokenInfoSchema>;
export type TokenResponse = z.infer<typeof TokenResponseSchema>;
