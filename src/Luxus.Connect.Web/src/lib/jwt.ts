import { jwtDecode, type JwtPayload } from 'jwt-decode';

export const decode = <T extends JwtPayload>(jwt: string): T | undefined => {
  try {
    return jwtDecode<T>(jwt);
  } catch {
    return undefined;
  }
};

export const isValid = (jwt: string): boolean => {
  const decoded = decode(jwt);
  return !!decoded?.exp && decoded.exp > Date.now() / 1000;
};
