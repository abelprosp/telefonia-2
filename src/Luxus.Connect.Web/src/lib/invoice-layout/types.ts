export type InvoiceLayoutTheme = {
  primaryColor: string;
  accentColor: string;
  borderColor: string;
  headerBackground: string;
  titleColor: string;
  textColor: string;
  tableHeaderBackground: string;
  borderRadius: number;
};

export type InvoiceLayoutBranding = {
  logoDataUrl: string;
  companyName: string;
  tagline: string;
  documentTitle: string;
};

export type InvoiceLayoutSectionToggle = {
  enabled: boolean;
  title?: string;
};

export type InvoiceLayoutSections = {
  userData: InvoiceLayoutSectionToggle;
  accountValue: InvoiceLayoutSectionToggle;
  billingDates: InvoiceLayoutSectionToggle;
  accountSummary: InvoiceLayoutSectionToggle;
  detailedConsumption: InvoiceLayoutSectionToggle;
};

export type InvoiceLayoutLabels = {
  name: string;
  address: string;
  phone: string;
  totalServices: string;
  discounts: string;
  billingPeriod: string;
  referenceMonth: string;
  dueDate: string;
  description: string;
  quantity: string;
  type: string;
  unitPrice: string;
  total: string;
  totalLabel: string;
};

export type InvoiceLayoutConfig = {
  theme: InvoiceLayoutTheme;
  branding: InvoiceLayoutBranding;
  sections: InvoiceLayoutSections;
  labels: InvoiceLayoutLabels;
};

export type InvoiceLayoutLineItem = {
  description: string;
  quantity: string;
  type: string;
  unitPrice: string;
  total: string;
};

export type InvoiceLayoutRenderData = {
  customerName: string;
  customerAddress: string;
  customerPhone: string;
  invoiceNumber: string;
  invoiceAmount: string;
  invoiceDueDate: string;
  invoiceIssueDate: string;
  referenceMonth: string;
  periodStart: string;
  periodEnd: string;
  servicesTotal: string;
  discounts: string;
  description?: string;
  lineItems?: InvoiceLayoutLineItem[];
};

export type InvoiceLayoutTemplate = {
  id: string;
  name: string;
  code: string;
  config_json: InvoiceLayoutConfig;
  active: boolean;
  created_at: string;
  updated_at: string;
};
