namespace Luxus.Connect.Contracts.Customers.Responses;

public sealed record CustomerPhoneLineLinkResponse(
    string CustomerId,
    string PhoneLineId,
    string PhoneLineNumber,
    string PhoneLineStatus,
    string LineClassification,
    DateOnly StartDate,
    DateOnly? EndDate,
    bool IsActive
);
