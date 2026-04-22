namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line050HService : LineRecord
    {
        public string PlanCode { get; set; } = default!;
        public string ServiceName { get; set; } = default!;
        public string Flags { get; set; } = default!;
        public decimal Total { get; set; } = default!;
        public int Quantity { get; set; }
        public string Unity { get; set; } = default!;
        public decimal Franchise { get; set; }
        public decimal Used { get; set; }
    }
}
