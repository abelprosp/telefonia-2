namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class InvoiceFranchiseLineDetail : LineRecord
    {
        public string ServiceDescription { get; set; } = default!;
        public decimal ServiceOrder { get; set; }
        public string PhoneNumber { get; set; } = default!;
        public string DetailSequence { get; set; } = default!;
        public decimal UsageAmount { get; set; }
        public string SectionReference { get; set; } = default!;
    }
}
