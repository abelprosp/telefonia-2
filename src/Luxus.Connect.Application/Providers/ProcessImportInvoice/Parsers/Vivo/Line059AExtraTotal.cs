namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line059AExtraTotal : LineRecord
    {
        public decimal TotalAmount { get; set; }
    }
}
