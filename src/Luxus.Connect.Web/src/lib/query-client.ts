import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 0,
      // Evita refetch de todas as listas a cada foco de janela.
      staleTime: 30_000
    },
    mutations: {
      retry: 0
    }
  }
});
