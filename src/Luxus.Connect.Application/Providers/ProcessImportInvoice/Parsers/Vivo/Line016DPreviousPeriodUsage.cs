namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line016DPreviousPeriodUsage : LineRecord
    {
        public decimal Quantity { get; set; }
        public decimal Amount { get; set; }
        public string TariffCode { get; set; } = default!;
    }
}
