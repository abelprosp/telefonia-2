namespace Luxus.Connect.Contracts.PhoneLines.Responses;

public sealed record PhoneLineCustomerLinkResponse(
    string PhoneLineId,
    string CustomerId,
    string CustomerName,
    string? CustomerDocument,
    DateOnly StartDate,
    DateOnly? EndDate,
    bool IsActive
);
