using Goal.Infra.Data.Auditing;
using Luxus.Connect.Infra.Crosscutting;

namespace Luxus.Connect.Infra.Data;

internal sealed class AppDbChangesInterceptor : AuditChangesInterceptor
{
    private readonly AppState appState;

    public AppDbChangesInterceptor(AppState appState)
    {
        this.appState = appState;
    }

    public override string GetCurrentUserId()
        => appState.User?.Username ?? appState.User?.Email ?? "system";
}
