import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import client from '@/lib/client';
import type { UserProfile } from '@/lib/auth-roles';

export type OrganizationUser = {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  full_name: string;
  profile: UserProfile;
  enabled: boolean;
};

export type CreateOrganizationUserInput = {
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  password: string;
  profile: UserProfile;
};

export type UpdateOrganizationUserInput = {
  profile?: UserProfile;
  enabled?: boolean;
  password?: string;
};

export const usersKeys = {
  all: ['organization-users'] as const,
  list: (search?: string) => ['organization-users', search ?? ''] as const
};

export function useOrganizationUsers(search?: string) {
  return useQuery({
    queryKey: usersKeys.list(search),
    queryFn: async () => {
      const { data } = await client<{ items: OrganizationUser[] }>({
        url: '/v1/users',
        method: 'GET',
        params: search ? { search } : undefined
      });
      return data.items ?? [];
    }
  });
}

export function useCreateOrganizationUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (input: CreateOrganizationUserInput) => {
      const { data } = await client<OrganizationUser>({
        url: '/v1/users',
        method: 'POST',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: usersKeys.all });
    }
  });
}

export function useUpdateOrganizationUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, ...input }: UpdateOrganizationUserInput & { id: string }) => {
      const { data } = await client<OrganizationUser>({
        url: `/v1/users/${id}`,
        method: 'PATCH',
        data: input
      });
      return data;
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: usersKeys.all });
    }
  });
}
