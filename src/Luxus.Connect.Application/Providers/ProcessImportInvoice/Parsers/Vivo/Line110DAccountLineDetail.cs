namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line110DAccountLineDetail : LineRecord
    {
        public string LineSequence { get; set; } = default!;
        public string PhoneNumber { get; set; } = default!;
        public string PlanName { get; set; } = default!;
        public decimal LineTotal { get; set; }
    }
}
