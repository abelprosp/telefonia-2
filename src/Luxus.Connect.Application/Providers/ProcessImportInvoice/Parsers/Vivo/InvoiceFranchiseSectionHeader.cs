namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class InvoiceFranchiseSectionHeader : LineRecord
    {
        public string SectionName { get; set; } = default!;
        public string TotalsPayload { get; set; } = default!;
        public string SectionReference { get; set; } = default!;
        public decimal SubscriberCount { get; set; }
        public string SectionSequence { get; set; } = default!;
    }
}
