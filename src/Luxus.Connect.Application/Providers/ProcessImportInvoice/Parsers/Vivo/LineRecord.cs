namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public abstract class LineRecord
    {
        public string AccountNumber { get; set; } = default!;
        public string BlockNumber { get; set; } = default!;
        public string BlockCode { get; set; } = default!;
        public string RecordType { get; set; } = default!;
    }
}
