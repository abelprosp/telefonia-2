namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line051DUsage : LineRecord
    {
        public string PlanCode { get; set; } = default!;
        public string ServiceName { get; set; } = default!;
        public string Unity { get; set; } = default!;
        public decimal Franchise { get; set; }
        public decimal Used { get; set; }
    }
}
