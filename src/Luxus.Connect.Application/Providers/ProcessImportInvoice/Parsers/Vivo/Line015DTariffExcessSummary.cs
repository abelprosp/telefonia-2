namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line015DTariffExcessSummary : LineRecord
    {
        public string RawPayload { get; set; } = default!;
    }
}
