using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Crosscutting.Errors;

public record ServiceUnavailableError(IEnumerable<Notification> Notifications)
    : AppError(ErrorType.ServiceUnavailable, Notifications)
{
    public ServiceUnavailableError(params Notification[] Notifications)
        : this(Notifications as IEnumerable<Notification>)
    {
    }
}