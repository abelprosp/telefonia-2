namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line050WServicesHeader : LineRecord
    {
        public string PlanCode { get; set; } = default!;
        public string Description { get; set; } = default!;
    }
}
