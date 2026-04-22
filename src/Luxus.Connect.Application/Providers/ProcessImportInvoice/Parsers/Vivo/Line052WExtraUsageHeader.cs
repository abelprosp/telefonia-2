namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line052WExtraUsageHeader : LineRecord
    {
        public string Description { get; set; } = default!;
        public string FooterFlags { get; set; } = default!;
    }
}
