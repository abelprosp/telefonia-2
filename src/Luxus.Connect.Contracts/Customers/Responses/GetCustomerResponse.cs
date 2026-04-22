using Goal.Infra.Crosscutting.Extensions;
using Luxus.Connect.Domain.Customers.Aggregates;

namespace Luxus.Connect.Contracts.Customers.Responses;

public sealed record GetCustomerResponse(
    string Id,
    bool Active,
    string Type,
    string Name,
    string CpfCnpj,
    string? StateRegistration,
    string? LegalName,
    DateOnly? BirthOrOpeningDate,
    IList<GetCustomerAddressResponse> Addresses)
{
    public static explicit operator GetCustomerResponse(Customer entity)
    {
        CustomerDocument? cpfCnpj = entity.Documents.SingleOrDefault(
            d => d.DocumentType is CustomerDocumentType.CPF or CustomerDocumentType.CNPJ);

        CustomerDocument? stateRegistration = entity.Documents.FirstOrDefault(
            d => d.DocumentType == CustomerDocumentType.STATE_REGISTRATION);

        string cpfCnpjNumber = cpfCnpj?.Number ?? string.Empty;

        return new GetCustomerResponse(
            entity.Id,
            entity.Active,
            entity.Type.GetDescription(),
            entity.Name,
            cpfCnpjNumber,
            stateRegistration?.Number,
            entity.LegalName,
            entity.BirthOrOpeningDate,
            [.. entity.Addresses.Select(address => (GetCustomerAddressResponse)address)]);
    }
}