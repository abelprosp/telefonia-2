namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class InvoiceFiscalNfcTotals : LineRecord
    {
        public string FiscalGroupCode { get; set; } = default!;
        public string Payload { get; set; } = default!;
    }
}
