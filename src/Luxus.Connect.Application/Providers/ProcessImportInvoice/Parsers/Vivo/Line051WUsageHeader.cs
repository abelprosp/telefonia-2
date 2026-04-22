namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line051WUsageHeader : LineRecord
    {
        public string Description { get; set; } = default!;
    }
}
