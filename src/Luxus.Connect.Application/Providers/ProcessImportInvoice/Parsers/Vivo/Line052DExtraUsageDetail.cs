namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line052DExtraUsageDetail : LineRecord
    {
        public string Description { get; set; } = default!;
        public decimal Quantity { get; set; }
        public decimal Amount { get; set; }
        public string ServiceCode { get; set; } = default!;
    }
}
