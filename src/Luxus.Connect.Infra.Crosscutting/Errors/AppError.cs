using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Crosscutting.Errors;

public abstract record AppError(ErrorType Type, IEnumerable<Notification> Notifications)
{
    public AppError(ErrorType Type, params Notification[] Notifications)
        : this(Type, Notifications as IEnumerable<Notification>)
    {
    }
}

