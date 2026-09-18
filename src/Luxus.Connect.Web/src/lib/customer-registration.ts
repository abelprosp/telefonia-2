import type { CreateCustomerAddressCommand, ListCustomerResponse } from '@/api';

export type CustomerRegistrationProfile = {
  service_unit?: string; ua?: string; show_ua?: boolean; account?: string; show_account?: boolean;
  birthplace?: string; sex?: string; identity_type?: string; identity_issued_on?: string;
  identity_issuer?: string; identity_state?: string; rg?: string; father_name?: string; mother_name?: string;
  contact_phone?: string; phone?: string; mobile?: string; emails?: string[]; contact_name?: string;
  marital_status?: string; spouse_cpf?: string; profession?: string; notes?: string; show_notes?: boolean;
};

export type CustomerRegistrationDraft = Omit<CustomerRegistrationProfile, 'emails'> & { emails_text: string };
export type RegisteredCustomer = ListCustomerResponse & {
  profile?: CustomerRegistrationProfile;
  addresses?: CreateCustomerAddressCommand[];
  billing_email?: string | null;
};

export const emptyCustomerAddress = (): CreateCustomerAddressCommand => ({ street: '', number: '', neighborhood: '', city: '', state: '', zip_code: '', complement: '', country: 'Brasil' });
export const registrationDraft = (profile?: CustomerRegistrationProfile): CustomerRegistrationDraft => {
  const { emails = [], ...rest } = profile ?? {};
  return { ...rest, emails_text: emails.join(', ') };
};
export function registrationPayload(draft: CustomerRegistrationDraft, isPj: boolean): CustomerRegistrationProfile {
  const { emails_text, ...rest } = draft;
  const profile = { ...rest, emails: emails_text.split(/[,;]+/).map(value => value.trim()).filter(Boolean) };
  return isPj ? { ...profile, birthplace: '', sex: '', identity_type: '', identity_issued_on: '', identity_issuer: '', identity_state: '', rg: '', father_name: '', mother_name: '', marital_status: '', spouse_cpf: '', profession: '' } : profile;
}
export const hasCustomerAddress = (address: CreateCustomerAddressCommand): boolean =>
  Boolean(address.street || address.number || address.neighborhood || address.city || address.state || address.zip_code || address.complement);

export const brazilianStates = ['AC','AL','AP','AM','BA','CE','DF','ES','GO','MA','MT','MS','MG','PA','PB','PR','PE','PI','RJ','RN','RS','RO','RR','SC','SP','SE','TO'];
