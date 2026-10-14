import { jwtDecode, type JwtPayload } from 'jwt-decode';

export const decode = <T extends JwtPayload>(jwt: string): T | undefined => {
  try {
    return jwtDecode<T>(jwt);
  } catch {
    return undefined;
  }
};
