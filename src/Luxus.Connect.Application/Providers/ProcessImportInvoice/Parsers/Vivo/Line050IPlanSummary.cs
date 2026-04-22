namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line050IPlanSummary : LineRecord
    {
        public string PlanCode { get; set; } = default!;
        public string PlanName { get; set; } = default!;
        public string Flags { get; set; } = default!;
        public decimal Subtotal { get; set; } = default!;
    }
}
