namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class InvoiceFiscalNfcItem : LineRecord
    {
        public string FiscalGroupCode { get; set; } = default!;
        public string ItemSubqualifier { get; set; } = default!;
        public string DetailPayload { get; set; } = default!;
    }
}
