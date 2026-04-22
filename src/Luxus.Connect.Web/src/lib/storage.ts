export const getItem = <T>(key: string, defaultValue?: T): T | undefined => {
  return (localStorage.getItem(key) as T) ?? defaultValue;
};

export const setItem = (key: string, value?: string) => {
  return localStorage.setItem(key, value ?? '');
};
