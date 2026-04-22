using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Http.Controllers;

public record ApiResponse
{
    public IEnumerable<Notification> Messages { get; init; }

    private ApiResponse(params Notification[] messages)
    {
        Messages = messages ?? [];
    }

    public static ApiResponse Fail(IEnumerable<Notification> messages)
        => Fail([.. messages]);

    public static ApiResponse Fail(params Notification[] messages)
        => new(messages);
}
