import type { InvoiceLayoutConfig, InvoiceLayoutRenderData } from './types';

export const DEFAULT_INVOICE_LAYOUT_CONFIG: InvoiceLayoutConfig = {
  theme: {
    primaryColor: '#4a4a4a',
    accentColor: '#00a0c6',
    borderColor: '#222222',
    headerBackground: '#ffffff',
    titleColor: '#1a1a1a',
    textColor: '#333333',
    tableHeaderBackground: '#f7f7f7',
    borderRadius: 12
  },
  branding: {
    logoDataUrl: '',
    companyName: 'LUXUS',
    tagline: 'SOLUÇÃO EM TELEFONIA',
    documentTitle: 'Detalhamento da Fatura'
  },
  sections: {
    userData: { enabled: true, title: 'Dados do Usuário' },
    accountValue: { enabled: true, title: 'VALOR DA SUA CONTA' },
    billingDates: { enabled: true },
    accountSummary: { enabled: true, title: 'Resumo da Conta' },
    detailedConsumption: { enabled: true, title: 'Consumo Detalhado' }
  },
  labels: {
    name: 'Nome:',
    address: 'Endereço:',
    phone: 'Número do telefone:',
    totalServices: 'Total Serviços:',
    discounts: 'Descontos:',
    billingPeriod: 'Período de faturamento:',
    referenceMonth: 'Mês de referência:',
    dueDate: 'Data de Vencimento:',
    description: 'Descrição',
    quantity: 'Quantidade',
    type: 'Tipo',
    unitPrice: 'Preço Unitário',
    total: 'Total',
    totalLabel: 'Total:'
  }
};

export const SAMPLE_INVOICE_LAYOUT_DATA: InvoiceLayoutRenderData = {
  customerName: 'REDOBRAI ARTUR ABEL SCHMITZ',
  customerAddress: 'Rua Venancio Aires, 400, Igrejinha - Lajeado - RS - CEP: 95.910-674',
  customerPhone: '(51) 99962-0231',
  invoiceNumber: 'FAT-001',
  invoiceAmount: 'R$ 0,00',
  invoiceDueDate: '10/06/2026',
  invoiceIssueDate: '16/04/2026',
  referenceMonth: '05/2026',
  periodStart: '16/04/2026',
  periodEnd: '15/05/2026',
  servicesTotal: 'R$ 19,99',
  discounts: 'R$ -24,99',
  lineItems: [
    {
      description: 'SMART ILIMITADO 3GB',
      quantity: '1',
      type: 'Mensal',
      unitPrice: 'R$ 19,99',
      total: 'R$ 19,99'
    },
    {
      description: 'PACOTE DE TORPEDOS 800 torpedos',
      quantity: '1',
      type: 'Mensal',
      unitPrice: 'R$ 0,00',
      total: 'R$ 0,00'
    },
    {
      description: 'Consumo',
      quantity: '1',
      type: 'Mensal',
      unitPrice: 'R$ 0,00',
      total: 'R$ 0,00'
    },
    {
      description: 'DESCONTO DE PLANO',
      quantity: '1',
      type: 'Mensal',
      unitPrice: 'R$ -24,99',
      total: 'R$ -24,99'
    }
  ]
};
