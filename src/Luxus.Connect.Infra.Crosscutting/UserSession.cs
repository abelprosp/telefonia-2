using System.Security.Claims;
using Luxus.Connect.Infra.Crosscutting.Extensions;

namespace Luxus.Connect.Infra.Crosscutting;

public class UserSession(ClaimsPrincipal principal)
{
    public DateTime ExpiresAt { get; private set; } = new DateTime(principal.GetExpiration(), DateTimeKind.Unspecified);
    public DateTime IssuedAt { get; private set; } = new DateTime(principal.GetIssuedAt(), DateTimeKind.Unspecified);
}