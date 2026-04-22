using System.Security.Claims;
using IdentityModel;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Microsoft.AspNetCore.Http;

namespace Luxus.Connect.Infra.Crosscutting;

public sealed class AppState
{
    public AppState(IHttpContextAccessor httpContextAccessor)
    {
        ClaimsPrincipal? principal = httpContextAccessor?.HttpContext?.User;

        if (principal is not null)
        {
            if (principal.TryGetClaimValue([JwtClaimTypes.Subject, ClaimTypes.NameIdentifier], out string _))
            {
                User = new AppUser(principal);
            }

            if (principal.TryGetClaimValue("organization", out string _))
            {
                Organization = new AppOrganization(principal);
            }

            Client = new AppClient(principal);
            Session  = new UserSession(principal);
            Scopes  = principal.GetScopes();
        }
    }

    public AppOrganization? Organization { get; init; }
    public AppUser? User { get; init; }
    public AppClient? Client { get; init; }
    public UserSession? Session { get; init; }
    public IEnumerable<string>? Scopes { get; init; }
}