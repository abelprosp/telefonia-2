import { useId } from 'react';
import type { CreateCustomerAddressCommand } from '@/api';
import { Field, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { brazilianStates, type CustomerRegistrationDraft } from '@/lib/customer-registration';

type Props = {
  value: CustomerRegistrationDraft;
  onChange: (value: CustomerRegistrationDraft) => void;
  address: CreateCustomerAddressCommand;
  onAddressChange: (address: CreateCustomerAddressCommand) => void;
  billingEmail: string;
  onBillingEmailChange: (value: string) => void;
  isPj: boolean;
  disabled?: boolean;
};

export function CustomerRegistrationFields({ value, onChange, address, onAddressChange, billingEmail, onBillingEmailChange, isPj, disabled }: Props) {
  const prefix = useId();
  const text = (key: keyof CustomerRegistrationDraft, label: string, type = 'text', placeholder?: string) => (
    <Field key={key}>
      <FieldLabel htmlFor={`${prefix}-${key}`}>{label}</FieldLabel>
      <Input id={`${prefix}-${key}`} type={type} value={String(value[key] ?? '')} disabled={disabled} maxLength={256} placeholder={placeholder}
        onChange={event => onChange({ ...value, [key]: event.target.value })} />
    </Field>
  );
  const select = (key: keyof CustomerRegistrationDraft, label: string, options: [string, string][]) => (
    <Field key={key}>
      <FieldLabel htmlFor={`${prefix}-${key}`}>{label}</FieldLabel>
      <select id={`${prefix}-${key}`} className="border-input bg-background h-10 w-full rounded-xl border px-3 text-sm" disabled={disabled} value={String(value[key] ?? '')}
        onChange={event => onChange({ ...value, [key]: event.target.value })}>
        <option value="">Selecione (opcional)</option>
        {options.map(([id, name]) => <option key={id} value={id}>{name}</option>)}
      </select>
    </Field>
  );
  const flag = (key: 'show_ua' | 'show_account' | 'show_notes', label: string) => (
    <label className="flex items-center gap-2 text-sm" key={key}>
      <input type="checkbox" checked={Boolean(value[key])} disabled={disabled} onChange={event => onChange({ ...value, [key]: event.target.checked })} />
      {label}
    </label>
  );
  const addressField = (key: keyof CreateCustomerAddressCommand, label: string, maxLength: number) => (
    <Field key={key}><FieldLabel htmlFor={`${prefix}-address-${key}`}>{label}</FieldLabel>
      <Input id={`${prefix}-address-${key}`} maxLength={maxLength} disabled={disabled} value={address[key] ?? ''}
        onChange={event => onAddressChange({ ...address, [key]: event.target.value })} />
    </Field>
  );
  return <div className="flex flex-col gap-6">
    <fieldset className="rounded-xl border p-4">
      <legend className="px-2 font-semibold">Atendimento e identificação interna</legend>
      <div className="grid gap-4 sm:grid-cols-2">
        {text('service_unit', 'Unidade de atendimento')}
        {text('contact_name', isPj ? 'Pessoa de contato na empresa' : 'Contato')}
        <div className="space-y-2">{text('ua', 'UA')}{flag('show_ua', 'Mostrar UA')}</div>
        <div className="space-y-2">{text('account', 'Conta')}{flag('show_account', 'Mostrar conta')}</div>
      </div>
    </fieldset>
    {!isPj && <fieldset className="rounded-xl border p-4">
      <legend className="px-2 font-semibold">Dados pessoais e documentos</legend>
      <div className="grid gap-4 sm:grid-cols-2">
        {text('birthplace', 'Naturalidade', 'text', 'Cidade e UF de nascimento')}
        {select('sex', 'Sexo', [['female','Feminino'],['male','Masculino'],['other','Outro'],['not_informed','Prefiro não informar']])}
        {select('identity_type','Tipo de documento de identidade',[['rg','RG'],['cnh','CNH'],['passport','Passaporte'],['rne','RNE / CRNM'],['other','Outro']])}
        {text('rg','RG / número do documento')}
        {text('identity_issued_on','Data de emissão','date')}
        {text('identity_issuer','Órgão expedidor')}
        {select('identity_state','UF do documento',brazilianStates.map(state => [state,state]))}
        {text('father_name','Nome do pai')}
        {text('mother_name','Nome da mãe')}
        {select('marital_status','Estado civil',[['single','Solteiro(a)'],['married','Casado(a)'],['divorced','Divorciado(a)'],['widowed','Viúvo(a)'],['separated','Separado(a)'],['civil_union','União estável']])}
        {text('spouse_cpf','CPF do cônjuge / companheiro','text','000.000.000-00')}
        {text('profession','Profissão')}
      </div>
    </fieldset>}
    <fieldset className="rounded-xl border p-4">
      <legend className="px-2 font-semibold">Telefones e e-mails</legend>
      <div className="grid gap-4 sm:grid-cols-2">
        {text('contact_phone','Telefone de contato','tel','(00) 00000-0000')}
        {text('phone','Telefone','tel','(00) 0000-0000')}
        {text('mobile','Celular','tel','(00) 00000-0000')}
        {text('emails_text','E-mails de contato','text','Separe os endereços por vírgula')}
        <Field className="sm:col-span-2"><FieldLabel htmlFor={`${prefix}-billing-email`}>E-mail de cobrança</FieldLabel>
          <Input id={`${prefix}-billing-email`} type="email" maxLength={256} value={billingEmail} disabled={disabled} onChange={event => onBillingEmailChange(event.target.value)} />
          <p className="text-muted-foreground text-xs">Endereço que receberá as faturas do cliente.</p>
        </Field>
      </div>
    </fieldset>
    <fieldset className="rounded-xl border p-4">
      <legend className="px-2 font-semibold">Endereço</legend>
      <div className="grid gap-4 sm:grid-cols-2">
        {addressField('zip_code','CEP',9)}
        {addressField('city','Cidade',128)}
        <Field><FieldLabel htmlFor={`${prefix}-state`}>UF</FieldLabel><select id={`${prefix}-state`} value={address.state} disabled={disabled} className="border-input bg-background h-10 w-full rounded-xl border px-3 text-sm" onChange={event => onAddressChange({ ...address, state: event.target.value })}>
          <option value="">Selecione</option>{brazilianStates.map(state => <option key={state} value={state}>{state}</option>)}
        </select></Field>
        {addressField('street','Endereço / logradouro',256)}
        {addressField('number','Número',16)}
        {addressField('neighborhood','Bairro',128)}
        {addressField('complement','Complemento',128)}
        {addressField('country','País',32)}
      </div>
    </fieldset>
    <Field>
      <FieldLabel htmlFor={`${prefix}-notes`}>Observação / atenção</FieldLabel>
      <Textarea id={`${prefix}-notes`} rows={4} maxLength={4000} disabled={disabled} value={value.notes ?? ''} onChange={event => onChange({ ...value, notes: event.target.value })} />
      {flag('show_notes','Mostrar observação')}
    </Field>
  </div>;
}
