using Luxus.Connect.Domain.PhoneLines.Aggregates;

namespace Luxus.Connect.Contracts.PhoneLines.Responses;

public sealed record GetPhoneLineServiceResponse(
    string Id,
    string PhoneLineId,
    string ProviderPlanServiceId,
    string Name,
    string Code,
    bool Recurring,
    decimal? Price,
    bool Active)
{
    public static explicit operator GetPhoneLineServiceResponse(PhoneLineService entity)
    {
        return new GetPhoneLineServiceResponse(
            entity.Id,
            entity.PhoneLineId,
            entity.ProviderPlanServiceId,
            entity.Name,
            entity.Code,
            entity.Recurring,
            entity.Price,
            entity.Active);
    }
}
