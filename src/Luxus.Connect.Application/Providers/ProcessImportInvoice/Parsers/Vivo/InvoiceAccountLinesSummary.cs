namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class InvoiceAccountLinesSummary : LineRecord
    {
        public decimal SubtotalServices { get; set; }
    }
}
