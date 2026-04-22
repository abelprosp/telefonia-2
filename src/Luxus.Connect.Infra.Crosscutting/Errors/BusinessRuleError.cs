using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Crosscutting.Errors;

public record BusinessRuleError(IEnumerable<Notification> Notifications)
    : AppError(ErrorType.BusinessRule, Notifications)
{
    public BusinessRuleError(params Notification[] Notifications)
        : this(Notifications as IEnumerable<Notification>)
    {
    }
}