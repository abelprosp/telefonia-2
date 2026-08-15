import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { client } from '@/lib/client';

export type UserProfile = {
  id: string;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  full_name: string;
  roles: string[];
  profile: string;
};

export type UpdateUserProfileInput = {
  first_name?: string;
  last_name?: string;
  email?: string;
  current_password?: string;
  new_password?: string;
};

export type CompanySettings = {
  company_name: string;
  trading_name: string;
  cnpj: string;
  state_registration: string;
  email: string;
  phone: string;
  website: string;
  zip_code: string;
  street: string;
  number: string;
  complement: string;
  neighborhood: string;
  city: string;
  state: string;
};

export type WhitelabelSettings = {
  app_name: string;
  app_slogan: string;
  logo_url: string;
  dark_logo_url: string;
  favicon_url: string;
  primary_color: string;
  support_email: string;
  support_phone: string;
  footer_text: string;
};

export type SystemSettings = {
  default_due_day: number;
  late_fee_percentage: number;
  interest_rate_monthly: number;
  days_before_due_reminder: number;
  days_after_due_reminder: number;
  auto_send_invoice_email: boolean;
  auto_send_collection_reminder: boolean;
};

export type OrganizationSettings = {
  organization_id: string;
  company: CompanySettings;
  whitelabel: WhitelabelSettings;
  system: SystemSettings;
  updated_at: string;
  updated_by?: string;
};

export async function fetchUserProfile(): Promise<UserProfile> {
  const res = await client<UserProfile>({
    url: '/v1/me',
    method: 'GET'
  });
  return res.data;
}

export async function updateUserProfile(data: UpdateUserProfileInput): Promise<UserProfile> {
  const res = await client<UserProfile, unknown, UpdateUserProfileInput>({
    url: '/v1/me',
    method: 'PATCH',
    data
  });
  return res.data;
}

export async function fetchOrganizationSettings(): Promise<OrganizationSettings> {
  const res = await client<OrganizationSettings>({
    url: '/v1/organization-settings',
    method: 'GET'
  });
  return res.data;
}

export async function updateCompanySettings(data: Partial<CompanySettings>): Promise<OrganizationSettings> {
  const res = await client<OrganizationSettings, unknown, Partial<CompanySettings>>({
    url: '/v1/company-settings',
    method: 'PUT',
    data
  });
  return res.data;
}

export async function updateWhitelabelSettings(data: Partial<WhitelabelSettings>): Promise<OrganizationSettings> {
  const res = await client<OrganizationSettings, unknown, Partial<WhitelabelSettings>>({
    url: '/v1/whitelabel-settings',
    method: 'PUT',
    data
  });
  return res.data;
}

export async function updateSystemSettings(data: Partial<SystemSettings>): Promise<OrganizationSettings> {
  const res = await client<OrganizationSettings, unknown, Partial<SystemSettings>>({
    url: '/v1/system-settings',
    method: 'PUT',
    data
  });
  return res.data;
}

// React Query Hooks
export function useUserProfileQuery() {
  return useQuery({
    queryKey: ['user-profile'],
    queryFn: fetchUserProfile,
    staleTime: 1000 * 60 * 5, // 5 min
    retry: 1
  });
}

export function useOrganizationSettingsQuery() {
  return useQuery({
    queryKey: ['organization-settings'],
    queryFn: fetchOrganizationSettings,
    staleTime: 1000 * 60 * 10, // 10 min
    retry: 1
  });
}

export function useUpdateUserProfileMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateUserProfile,
    onSuccess: (data) => {
      queryClient.setQueryData(['user-profile'], data);
    }
  });
}

export function useUpdateCompanySettingsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateCompanySettings,
    onSuccess: (data) => {
      queryClient.setQueryData(['organization-settings'], data);
    }
  });
}

export function useUpdateWhitelabelSettingsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateWhitelabelSettings,
    onSuccess: (data) => {
      queryClient.setQueryData(['organization-settings'], data);
    }
  });
}

export function useUpdateSystemSettingsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateSystemSettings,
    onSuccess: (data) => {
      queryClient.setQueryData(['organization-settings'], data);
    }
  });
}
