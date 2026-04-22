namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line052EExtraLocation : LineRecord
    {
        public string Location { get; set; } = default!;
    }
}
