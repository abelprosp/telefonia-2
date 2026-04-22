namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line010DHeader : LineRecord
    {
        public string ReferenceMonth { get; set; } = default!;
        public DateOnly IssueDate { get; set; } = default!;
        public DateOnly DueDate { get; set; } = default!;
        public DateOnly BillingStartDate { get; set; } = default!;
        public DateOnly BillingEndDate { get; set; } = default!;
        public decimal SubtotalServices { get; set; }
        public decimal SubtotalUsageExceeded { get; set; }
        public decimal TotalAmount { get; set; }
        public string FiscalReferenceCode { get; set; } = default!;
    }
}
