import { useMemo } from 'react';

import { useAuth } from 'react-oidc-context';

import { decode } from '@/lib/jwt';
import { type TokenPayload } from '@/types/auth';

export type UserProfile = 'master' | 'employee' | 'financial' | 'partner' | 'user';

export type AuthRoleState = {
  roles: string[];
  isMaster: boolean;
  isEmployee: boolean;
  isFinancial: boolean;
  isPartner: boolean;
  isInternalStaff: boolean;
  isPartnerOnly: boolean;
  canAccessOperations: boolean;
  canAccessFinance: boolean;
  canManageUsers: boolean;
  profile: UserProfile;
};

function rolesFromAccessToken(token: string | undefined): string[] {
  if (!token) {
    return [];
  }

  const payload = decode<TokenPayload & { roles?: string[] }>(token);
  if (!payload) {
    return [];
  }

  const fromClaim = payload.roles ?? [];
  const fromRealm = payload.realm_access?.roles ?? [];
  const fromClient = Object.values(payload.resource_access ?? {}).flatMap(
    (entry) => entry.roles ?? []
  );

  return [...new Set([...fromClaim, ...fromRealm, ...fromClient])];
}

export function resolveProfile(roles: string[]): UserProfile {
  if (roles.includes('master') || roles.includes('admin')) return 'master';
  if (roles.includes('financial')) return 'financial';
  if (roles.includes('employee')) return 'employee';
  if (roles.includes('partner')) return 'partner';
  return 'user';
}

export function profileLabel(profile: UserProfile): string {
  const map: Record<UserProfile, string> = {
    master: 'Master',
    employee: 'Funcionário',
    financial: 'Financeiro',
    partner: 'Parceiro',
    user: 'Usuário'
  };
  return map[profile];
}

export function useAuthRoles(): AuthRoleState {
  const { user } = useAuth();

  return useMemo(() => {
    const roles = rolesFromAccessToken(user?.access_token);
    const isMaster = roles.includes('master') || roles.includes('admin');
    const isEmployee = roles.includes('employee');
    const isFinancial = roles.includes('financial');
    const isPartner = roles.includes('partner');
    const isInternalStaff = isMaster || isEmployee || isFinancial;
    const profile = resolveProfile(roles);

    return {
      roles,
      isMaster,
      isEmployee,
      isFinancial,
      isPartner,
      isInternalStaff,
      isPartnerOnly: isPartner && !isInternalStaff,
      canAccessOperations: isMaster || isEmployee,
      canAccessFinance: isMaster || isFinancial,
      canManageUsers: isMaster,
      profile
    };
  }, [user?.access_token]);
}

export function roleLabel(roles: Pick<AuthRoleState, 'profile'>) {
  return profileLabel(roles.profile);
}

export const PROFILE_OPTIONS: { value: UserProfile; label: string; description: string }[] = [
  {
    value: 'master',
    label: 'Master',
    description: 'Acesso total ao sistema e gestão de usuários'
  },
  {
    value: 'employee',
    label: 'Funcionário',
    description: 'Operação básica: clientes, linhas, faturamento e vendas (sem financeiro)'
  },
  {
    value: 'financial',
    label: 'Financeiro',
    description: 'Controle financeiro completo: contas, comissões e relatórios'
  },
  {
    value: 'partner',
    label: 'Parceiro',
    description: 'Portal do parceiro: clientes, linhas e vendas da carteira'
  }
];
