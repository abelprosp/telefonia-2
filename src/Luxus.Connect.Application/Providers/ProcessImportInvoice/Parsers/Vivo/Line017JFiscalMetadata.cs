namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line017JFiscalMetadata : LineRecord
    {
        public string Payload { get; set; } = default!;
    }
}
