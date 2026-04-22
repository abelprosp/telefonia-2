using Goal.Domain.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Extensions;

namespace Luxus.Connect.Domain.Customers.Aggregates;

public class Customer : Entity
{
    protected Customer()
        : base()
    {
    }

    private Customer(string organizationId, string name, CustomerType type)
        : this()
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(organizationId);

        OrganizationId = organizationId;
        Name = name;
        Type = type;
    }

    public string OrganizationId { get; private set; } = default!;
    public CustomerType Type { get; private set; }
    public string Name { get; private set; } = default!;
    public string? LegalName { get; private set; }
    public DateOnly? BirthOrOpeningDate { get; private set; }
    public string? ResponsibleSalespersonUserId { get; private set; }
    public bool Active { get; private set; } = true;
    public IEnumerable<CustomerDocument> Documents { get; private set; } = Enumerable.Empty<CustomerDocument>().ToList();
    public IEnumerable<CustomerAddress> Addresses { get; private set; } = Enumerable.Empty<CustomerAddress>().ToList();
    public IEnumerable<CustomerProviderLink> ProviderLinks { get; private set; } = Enumerable.Empty<CustomerProviderLink>().ToList();
    public IEnumerable<PhoneLineCustomerLink> PhoneLineLinks { get; private set; } = Enumerable.Empty<PhoneLineCustomerLink>().ToList();
    public IEnumerable<CustomerProcessingMonthManualRelease> CustomerProcessingMonthManualReleases { get; private set; } = Enumerable.Empty<CustomerProcessingMonthManualRelease>().ToList();
    public IEnumerable<CustomerAttachment> Attachments { get; private set; } = Enumerable.Empty<CustomerAttachment>().ToList();

    public void AddDocument(CustomerDocument document)
    {
        Documents = Documents
            .Append(document)
            .ToList();
    }

    public void UpdateName(string name)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        Name = name;
    }

    public void UpdateLegalName(string? legalName)
        => LegalName = legalName;

    public void UpdateBirthOrOpeningDate(DateOnly? date)
        => BirthOrOpeningDate = date;

    public void SetResponsibleSalespersonUserId(string? userId)
        => ResponsibleSalespersonUserId = string.IsNullOrWhiteSpace(userId) ? null : userId.Trim();

    public void Inactivate()
        => Active = false;

    public void Reactivate()
        => Active = true;

    public void UpdateStateRegistration(string? stateRegistration)
    {
        if (string.IsNullOrWhiteSpace(stateRegistration))
            return;

        CustomerDocument? existing = Documents
            .SingleOrDefault(d => d.DocumentType == CustomerDocumentType.STATE_REGISTRATION);

        if (existing is not null)
            existing.UpdateNumber(stateRegistration);
        else
            AddDocument(CustomerDocument.Create(this, CustomerDocumentType.STATE_REGISTRATION, stateRegistration));
    }

    public void AddAddress(
        string street,
        string number,
        string neighborhood,
        string city,
        string state,
        string zipCode,
        string? complement,
        string country)
    {
        Addresses = Addresses
            .Append(CustomerAddress.Create(this, street, number, neighborhood, city, state, zipCode, complement, country))
            .ToList();
    }

    public void UpdateOrAddPrimaryAddress(
        string street,
        string number,
        string neighborhood,
        string city,
        string state,
        string zipCode,
        string? complement,
        string country)
    {
        CustomerAddress? first = Addresses.FirstOrDefault();

        if (first is null)
        {
            AddAddress(street, number, neighborhood, city, state, zipCode, complement, country);
            return;
        }

        first.Replace(street, number, neighborhood, city, state, zipCode, complement, country);
    }

    public static Customer Create(string organizationId, Provider provider, string name, string taxId)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(organizationId);
        ArgumentNullException.ThrowIfNull(provider);
        ArgumentException.ThrowIfNullOrWhiteSpace(name);
        ArgumentException.ThrowIfNullOrWhiteSpace(taxId);

        string normalizedTaxId = taxId.NormalizeDigitsOnly();

        if (normalizedTaxId.Length is not (11 or 14))
            throw new ArgumentException("Documento deve ter 11 (CPF) ou 14 (CNPJ) dígitos.");

        CustomerType type = normalizedTaxId.Length == 14
            ? CustomerType.PJ
            : CustomerType.PF;

        var customer = new Customer(organizationId, name, type);

        CustomerDocumentType docType = type == CustomerType.PJ
            ? CustomerDocumentType.CNPJ
            : CustomerDocumentType.CPF;

        customer.AddDocument(CustomerDocument.Create(customer, docType, normalizedTaxId));
        customer.AddProviderLink(provider, DateOnly.FromDateTime(DateTime.UtcNow));

        return customer;
    }

    public bool HasActiveProvider(string providerId)
        => ProviderLinks.Any(l => l.ProviderId == providerId && l.EndDate is null);

    public void AddProviderLink(Provider provider, DateOnly startDate)
    {
        ArgumentNullException.ThrowIfNull(provider);

        if (HasActiveProvider(provider.Id))
            return;

        ProviderLinks = ProviderLinks
            .Append(CustomerProviderLink.Create(this, provider, startDate))
            .ToList();
    }

    public string? GetCpfOrCnpj()
    {
        return Documents
            .FirstOrDefault(d => d.DocumentType is CustomerDocumentType.CPF or CustomerDocumentType.CNPJ)?.Number;
    }
}
