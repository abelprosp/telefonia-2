import { type AxiosError, type AxiosResponse } from 'axios';

/** Formato de erro da API (OpenAPI `ApiResponse`). */
export type ApiNotification = {
  code: string;
  message: string;
  param?: string | null;
};

export type ApiErrorBody = {
  messages?: ApiNotification[];
};

const isRecord = (v: unknown): v is Record<string, unknown> =>
  typeof v === 'object' && v !== null;

const isApiNotification = (v: unknown): v is ApiNotification =>
  isRecord(v) && typeof v.code === 'string' && typeof v.message === 'string';

export const parseApiErrorBody = (data: unknown): ApiNotification[] => {
  if (!isRecord(data)) {
    return [];
  }
  const raw = data.messages;
  if (!Array.isArray(raw)) {
    return [];
  }
  return raw.filter(isApiNotification);
};

export const formatApiMessages = (messages: readonly ApiNotification[]) =>
  messages.map((m) => m.message).filter(Boolean);

/**
 * Erro HTTP da API Luxus.Connect com mensagens estruturadas quando disponíveis.
 */
export class ApiHttpError extends Error {
  readonly status: number;
  readonly messages: ApiNotification[];
  readonly rawBody: unknown;

  constructor(
    status: number,
    messages: ApiNotification[],
    rawBody: unknown,
    fallbackMessage: string
  ) {
    super(
      messages.length > 0
        ? formatApiMessages(messages).join('\n')
        : fallbackMessage
    );
    this.name = 'ApiHttpError';
    this.status = status;
    this.messages = [...messages];
    this.rawBody = rawBody;
  }
}

export const isApiHttpError = (e: unknown): e is ApiHttpError =>
  e instanceof ApiHttpError;

const fallbackStatusText = (status: number) => {
  switch (status) {
    case 400:
      return 'Requisição inválida.';
    case 401:
      return 'Não autorizado.';
    case 403:
      return 'Acesso negado.';
    case 404:
      return 'Recurso não encontrado.';
    case 409:
      return 'Conflito ao salvar.';
    default:
      return `Erro HTTP ${status}.`;
  }
};

export const toApiHttpError = (response: AxiosResponse<unknown>) => {
  const data = response.data;
  const messages = parseApiErrorBody(data);
  return new ApiHttpError(
    response.status,
    messages,
    data,
    fallbackStatusText(response.status)
  );
};

export const getErrorMessage = (e: unknown): string => {
  if (isApiHttpError(e)) {
    return e.message;
  }
  if (typeof e === 'object' && e !== null && 'message' in e) {
    const m = (e as { message?: unknown }).message;
    if (typeof m === 'string' && m.length > 0) {
      return m;
    }
  }
  return 'Ocorreu um erro inesperado.';
};

export const isAxiosError = (e: unknown): e is AxiosError =>
  typeof e === 'object' &&
  e !== null &&
  'isAxiosError' in e &&
  (e as AxiosError).isAxiosError === true;
