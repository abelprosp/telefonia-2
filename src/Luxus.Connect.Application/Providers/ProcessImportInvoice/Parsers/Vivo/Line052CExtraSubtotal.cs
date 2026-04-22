namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line052CExtraSubtotal : LineRecord
    {
        public decimal Amount { get; set; }
    }
}
