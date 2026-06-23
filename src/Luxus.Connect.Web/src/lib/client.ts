import type {
  AxiosError,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig
} from 'axios';
import axios from 'axios';

import { env } from '@/env';

import { isAxiosError, toApiHttpError } from './api-error';

let getAccessTokenFromAuth: () => string | null = () => null;
let onUnauthorized: (() => void) | undefined;

export const registerAuthTokenGetter = (getter: () => string | null) => {
  getAccessTokenFromAuth = getter;
  return () => {
    getAccessTokenFromAuth = () => null;
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
    | 'arraybuffer'
    | 'blob'
    | 'document'
    | 'json'
    | 'text'
    | 'stream';
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

export const api = axios.create(getConfig());

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const headers = config.headers;
  const existing =
    typeof headers.get === 'function'
      ? headers.get('Authorization')
      : (headers as { Authorization?: string }).Authorization;

  if (existing) {
    return config;
  }

  const token = getAccessTokenFromAuth();

  if (token) {
    if (typeof headers.set === 'function') {
      headers.set('Authorization', `Bearer ${token}`);
    } else {
      (headers as { Authorization?: string }).Authorization = `Bearer ${token}`;
    }
  }

  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    const response = error.response;

    if (response) {
      const hadToken = Boolean(getAccessTokenFromAuth());

      if (response.status === 401 && hadToken) {
        onUnauthorized?.();
      }

      throw toApiHttpError(response);
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
