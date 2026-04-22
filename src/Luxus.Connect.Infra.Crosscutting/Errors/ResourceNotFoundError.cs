using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Crosscutting.Errors;

public record ResourceNotFoundError(IEnumerable<Notification> Notifications)
    : AppError(ErrorType.ResourceNotFound, Notifications)
{
    public ResourceNotFoundError(params Notification[] Notifications)
        : this(Notifications as IEnumerable<Notification>)
    {
    }
}