import type { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import axios from 'axios';

import { env } from '@/env';

import { isAxiosError, toApiHttpError } from './api-error';

let getRolesFromAuth: () => string[] | undefined = () => undefined;
let onUnauthorized: (() => void) | undefined;

export const getCurrentRoles = () => getRolesFromAuth();
export const registerAuthRolesGetter = (getter: () => string[] | undefined) => {
  getRolesFromAuth = getter;
  return () => {
    getRolesFromAuth = () => undefined;
  };
};

export const registerUnauthorizedHandler = (handler: () => void) => {
  onUnauthorized = handler;
  return () => {
    onUnauthorized = undefined;
  };
};

export type RequestConfig<TData = unknown> = {
  baseURL?: string;
  url?: string;
  method?: 'GET' | 'PUT' | 'PATCH' | 'POST' | 'DELETE' | 'OPTIONS' | 'HEAD';
  params?: unknown;
  data?: TData | FormData;
  responseType?:
    'arraybuffer' | 'blob' | 'document' | 'json' | 'text' | 'stream';
  signal?: AbortSignal;
  validateStatus?: (status: number) => boolean;
  headers?: AxiosRequestConfig['headers'];
  paramsSerializer?: AxiosRequestConfig['paramsSerializer'];
};

/**
 * Subset of AxiosResponse
 */
export type ResponseConfig<TData = unknown> = {
  data: TData;
  status: number;
  statusText: string;
  headers: AxiosResponse['headers'];
};

export type ResponseErrorConfig<TError = unknown> = AxiosError<TError>;

export type Client = <TResponseData, _TError = unknown, TRequestData = unknown>(
  config: RequestConfig<TRequestData>
) => Promise<ResponseConfig<TResponseData>>;

let _config: Partial<RequestConfig> = {
  baseURL: env.VITE_API_URL,
  headers: {
    ['Content-Type']: 'application/json'
  }
};

export const getConfig = () => _config;

export const setConfig = (config: RequestConfig) => {
  _config = config;
  return getConfig();
};

export const mergeConfig = <T extends RequestConfig>(
  ...configs: Array<Partial<T>>
): Partial<T> => {
  return configs.reduce<Partial<T>>((merged, config) => {
    return {
      ...merged,
      ...config,
      headers: {
        ...merged.headers,
        ...config.headers
      }
    };
  }, {});
};

export const api = axios.create({ ...getConfig(), withCredentials: true });

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    const response = error.response;

    if (response) {
      const hadToken = Boolean(getRolesFromAuth());

      if (response.status === 401 && hadToken) {
        onUnauthorized?.();
      }

      const apiErr = toApiHttpError(response);
      console.warn(
        `[API ${response.status}] ${response.config.method?.toUpperCase()} ${response.config.url}:`,
        apiErr.message,
        // Never log the raw response body in production (may contain PII).
        import.meta.env.DEV ? response.data : undefined
      );
      throw apiErr;
    }

    if (isAxiosError(error) && error.code === 'ERR_NETWORK') {
      throw new Error(
        'Sem conexão com o servidor. Verifique a rede e tente de novo.'
      );
    }

    return Promise.reject(error);
  }
);

export const client = async <
  TResponseData,
  TError = unknown,
  TRequestData = unknown
>(
  config: RequestConfig<TRequestData>
): Promise<ResponseConfig<TResponseData>> => {
  const _config = mergeConfig(getConfig(), config);

  return api
    .request<TResponseData, ResponseConfig<TResponseData>>(_config)
    .catch((e: AxiosError<TError>) => {
      throw e;
    });
};

client.getConfig = getConfig;
client.setConfig = setConfig;

export default client;
