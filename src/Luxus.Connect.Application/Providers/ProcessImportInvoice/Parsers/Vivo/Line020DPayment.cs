namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line020DPayment : LineRecord
    {
        public string DigitableLine { get; set; } = default!;
        public string PixQrCode { get; set; } = default!;
    }
}
