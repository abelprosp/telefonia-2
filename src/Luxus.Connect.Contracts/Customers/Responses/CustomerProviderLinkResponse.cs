namespace Luxus.Connect.Contracts.Customers.Responses;

public sealed record CustomerProviderLinkResponse(
    string CustomerId,
    string ProviderId,
    string ProviderName,
    DateOnly StartDate,
    DateOnly? EndDate,
    bool IsActive
);
