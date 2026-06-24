import type { InvoiceLayoutConfig, InvoiceLayoutLineItem, InvoiceLayoutRenderData } from './types';

function esc(value: string) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;');
}

function row(label: string, value: string) {
  return `<div style="margin-bottom:4px;"><strong>${esc(label)}</strong> ${esc(value)}</div>`;
}

function amountRow(label: string, value: string, highlight: boolean) {
  const weight = highlight ? '700' : 'normal';
  const size = highlight ? '15px' : '13px';
  return `<tr><td style="padding:4px 0;font-weight:${weight};font-size:${size};">${esc(label)}</td><td style="padding:4px 0;text-align:right;font-weight:${weight};font-size:${size};">${esc(value)}</td></tr>`;
}

function defaultConsumption(theme: InvoiceLayoutConfig['theme']) {
  const bg = theme.tableHeaderBackground;
  return `<div style="font-size:12px;">
<div style="font-weight:700;margin-bottom:6px;">PACOTE DE TORPEDOS CONSUMIDOS: 0</div>
<table role="presentation" width="100%" cellspacing="0" cellpadding="4" style="border-collapse:collapse;">
<tr style="background:${bg};font-weight:700;"><td>CHAMADAS LOCAIS</td><td align="right">R$</td></tr>
<tr><td style="padding-left:12px;">Fixo</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Outras Operadoras</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Vivo</td><td align="right">0,00</td></tr>
<tr style="background:${bg};font-weight:700;"><td>CHAMADAS ESTADUAIS</td><td align="right">R$</td></tr>
<tr><td style="padding-left:12px;">Fixo</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Outras Operadoras</td><td align="right">0,00</td></tr>
<tr><td style="padding-left:12px;">Móvel Vivo</td><td align="right">0,00</td></tr>
<tr style="background:${bg};font-weight:700;"><td>TORPEDOS</td><td align="right">R$</td></tr>
<tr><td style="padding-left:12px;">Móvel Vivo</td><td align="right">0,00</td></tr>
</table></div>`;
}

function renderLineItemsTable(
  config: InvoiceLayoutConfig,
  data: InvoiceLayoutRenderData
) {
  const { theme, labels } = config;
  let items: InvoiceLayoutLineItem[] = data.lineItems ?? [];
  if (items.length === 0 && data.description) {
    items = [
      {
        description: data.description,
        quantity: '1',
        type: 'Mensal',
        unitPrice: data.invoiceAmount,
        total: data.invoiceAmount
      }
    ];
  }

  const rows = items
    .map(
      (item) =>
        `<tr><td>${esc(item.description)}</td><td align="center">${esc(item.quantity)}</td><td align="center">${esc(item.type)}</td><td align="right">${esc(item.unitPrice)}</td><td align="right">${esc(item.total)}</td></tr>`
    )
    .join('');

  return `<table role="presentation" width="100%" cellspacing="0" cellpadding="6" style="border-collapse:collapse;font-size:12px;">
<thead><tr style="background:${theme.tableHeaderBackground};">
<th align="left">${esc(labels.description)}</th>
<th>${esc(labels.quantity)}</th>
<th>${esc(labels.type)}</th>
<th align="right">${esc(labels.unitPrice)}</th>
<th align="right">${esc(labels.total)}</th>
</tr></thead>
<tbody>${rows}
<tr><td colspan="4" align="right" style="font-weight:700;padding-top:8px;">${esc(labels.totalLabel)}</td><td align="right" style="font-weight:700;padding-top:8px;">${esc(data.invoiceAmount)}</td></tr>
</tbody></table>`;
}

export function renderInvoiceLayoutHtml(
  config: InvoiceLayoutConfig,
  data: InvoiceLayoutRenderData
): string {
  const t = config.theme;
  const l = config.labels;
  const radius = t.borderRadius > 0 ? t.borderRadius : 12;
  const box = `border:1px solid ${t.borderColor};border-radius:${radius}px;padding:14px 16px;margin-bottom:12px;background:${t.headerBackground};`;

  let html = `<div style="font-family:Arial,Helvetica,sans-serif;max-width:820px;margin:0 auto;color:${t.textColor};font-size:13px;line-height:1.45;">`;

  html += `<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="margin-bottom:12px;"><tr>`;
  html += `<td width="50%" style="vertical-align:top;padding-right:6px;"><div style="${box}text-align:center;">`;
  if (config.branding.logoDataUrl) {
    html += `<img src="${config.branding.logoDataUrl}" alt="Logo" style="max-height:72px;max-width:100%;object-fit:contain;margin-bottom:8px;" />`;
  }
  html += `<div style="font-size:22px;font-weight:700;color:${t.primaryColor};letter-spacing:1px;">${esc(config.branding.companyName)}</div>`;
  html += `<div style="font-size:11px;color:${t.textColor};">${esc(config.branding.tagline)}</div></div></td>`;
  html += `<td width="50%" style="vertical-align:top;padding-left:6px;"><div style="${box}text-align:center;height:100%;">`;
  html += `<h1 style="margin:0;font-size:26px;color:${t.titleColor};">${esc(config.branding.documentTitle)}</h1></div></td></tr></table>`;

  if (config.sections.userData.enabled) {
    html += `<div style="${box}">`;
    if (config.sections.userData.title) {
      html += `<div style="font-weight:700;margin-bottom:8px;">${esc(config.sections.userData.title)}</div>`;
    }
    html += row(l.name, data.customerName);
    html += row(l.address, data.customerAddress);
    html += row(l.phone, data.customerPhone);
    html += `</div>`;
  }

  if (config.sections.accountValue.enabled) {
    html += `<div style="${box}"><table role="presentation" width="100%" cellspacing="0" cellpadding="0">`;
    html += amountRow(config.sections.accountValue.title ?? '', data.invoiceAmount, true);
    html += amountRow(l.totalServices, data.servicesTotal, false);
    html += amountRow(l.discounts, data.discounts, false);
    html += `</table></div>`;
  }

  if (config.sections.billingDates.enabled) {
    html += `<div style="${box}">`;
    html += row(l.billingPeriod, `${data.periodStart} a ${data.periodEnd}`);
    html += row(l.referenceMonth, data.referenceMonth);
    html += row(l.dueDate, data.invoiceDueDate);
    html += `</div>`;
  }

  if (config.sections.accountSummary.enabled) {
    html += `<div style="${box}">`;
    if (config.sections.accountSummary.title) {
      html += `<div style="font-weight:700;margin-bottom:8px;">${esc(config.sections.accountSummary.title)}</div>`;
    }
    html += renderLineItemsTable(config, data);
    html += `</div>`;
  }

  if (config.sections.detailedConsumption.enabled) {
    html += `<div style="${box}">`;
    if (config.sections.detailedConsumption.title) {
      html += `<div style="font-weight:700;margin-bottom:8px;">${esc(config.sections.detailedConsumption.title)}</div>`;
    }
    html += defaultConsumption(t);
    html += `</div>`;
  }

  html += `</div>`;
  return wrapHtmlDocument(html);
}

export function wrapHtmlDocument(body: string) {
  const trimmed = body.trim();
  if (!trimmed) {
    return '<!DOCTYPE html><html lang="pt-BR"><head><meta charset="UTF-8"></head><body></body></html>';
  }
  if (/<html/i.test(trimmed)) {
    if (/charset/i.test(trimmed)) return trimmed;
    return trimmed.replace(
      /<head>/i,
      '<head><meta charset="UTF-8"><meta http-equiv="Content-Type" content="text/html; charset=UTF-8">'
    );
  }
  return `<!DOCTYPE html><html lang="pt-BR"><head><meta charset="UTF-8"><meta http-equiv="Content-Type" content="text/html; charset=UTF-8"></head><body style="margin:0;padding:16px;background:#ffffff;">${trimmed}</body></html>`;
}
