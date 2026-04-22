using System.Security.Claims;
using Luxus.Connect.Infra.Crosscutting.Extensions;

namespace Luxus.Connect.Infra.Crosscutting;

public sealed class AppClient(ClaimsPrincipal principal)
{
    public string ClientId { get; } = principal.GetClientId();
}