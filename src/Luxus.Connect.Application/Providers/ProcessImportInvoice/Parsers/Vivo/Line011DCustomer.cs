namespace Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;

public static partial class VivoTextInvoiceParser
{
    public class Line011DCustomer : LineRecord
    {
        public string Name { get; set; } = default!;
        public string LegalName { get; set; } = default!;
        public string Document { get; set; } = default!;
        public string Street { get; set; } = default!;
        public string Number { get; set; } = default!;
        public string Neighborhood { get; set; } = default!;
        public string ZipCode { get; set; } = default!;
        public string City { get; set; } = default!;
        public string State { get; set; } = default!;
        public string Country { get; set; } = default!;
        public string StateRegistration { get; set; } = default!;
    }
}
