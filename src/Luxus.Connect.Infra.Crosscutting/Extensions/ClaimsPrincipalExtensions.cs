using System.Security.Claims;
using System.Text.Json;
using IdentityModel;

namespace Luxus.Connect.Infra.Crosscutting.Extensions;

public static class ClaimsPrincipalExtensions
{
    public static string? GetClaimValue(this ClaimsPrincipal principal, params string[] claimTypes)
    {
        Claim? claim = principal
            .Claims
            .Where(p => claimTypes.Contains(p.Type))
            .FirstOrDefault();

        return claim?.Value;
    }

    public static IEnumerable<string> GetClaimValues(this ClaimsPrincipal principal, params string[] claimTypes)
    {
        return principal
            .Claims
            .Where(p => claimTypes.Contains(p.Type))
            .Select(p => p.Value)
            .ToHashSet();
    }

    public static bool TryGetClaimValue(this ClaimsPrincipal principal, string claimType, out string claimValue)
        => principal.TryGetClaimValue([claimType], out claimValue);

    public static bool TryGetClaimValue(this ClaimsPrincipal principal, string[] claimTypes, out string claimValue)
    {
        claimValue = default!;
        string? value = principal.GetClaimValue(claimTypes);

        if (!string.IsNullOrWhiteSpace(value))
        {
            claimValue = value;
            return true;
        }

        return false;
    }

    public static T GetClaimValueAs<T>(this ClaimsPrincipal principal, params string[] claimTypes)
        where T : struct
    {
        if (!principal.TryGetClaimValue(claimTypes, out string claimValue))
        {
            return default;
        }

        try
        {
            return (T)Convert.ChangeType(claimValue, typeof(T));
        }
        catch
        {
            return default;
        }
    }

    public static string GetUserId(this ClaimsPrincipal principal)
        => principal.GetClaimValue(JwtClaimTypes.Subject, ClaimTypes.NameIdentifier)!;

    public static Dictionary<string, T>? GetOrganization<T>(this ClaimsPrincipal principal)
    {
        string? claim = principal.GetClaimValue("organization");

        if (string.IsNullOrWhiteSpace(claim))
        {
            throw new InvalidOperationException("Organization claim is missing.");
        }

        return JsonSerializer.Deserialize<Dictionary<string, T>>(claim);
    }

    public static string GetEmail(this ClaimsPrincipal principal)
        => principal?.GetClaimValue(JwtClaimTypes.Email, ClaimTypes.Email)!;

    public static string GetName(this ClaimsPrincipal principal)
        => principal?.GetClaimValue(JwtClaimTypes.Name, ClaimTypes.Name)!;

    public static string GetUsername(this ClaimsPrincipal principal)
        => principal?.GetClaimValue(JwtClaimTypes.PreferredUserName)!;

    public static string? GetGivenName(this ClaimsPrincipal principal)
        => principal?.GetClaimValue(JwtClaimTypes.GivenName, ClaimTypes.GivenName);

    public static string? GetFamilyName(this ClaimsPrincipal principal)
        => principal?.GetClaimValue(JwtClaimTypes.FamilyName);

    public static IEnumerable<string> GetRoles(this ClaimsPrincipal principal)
        => principal?.GetClaimValues(JwtClaimTypes.Role, ClaimTypes.Role) ?? [];

    public static IEnumerable<string> GetScopes(this ClaimsPrincipal principal)
        => principal?.GetClaimValues(JwtClaimTypes.Scope) ?? [];

    public static string GetClientId(this ClaimsPrincipal principal)
        => principal?.GetClaimValue(JwtClaimTypes.ClientId, JwtClaimTypes.AuthorizedParty)!;

    public static long GetExpiration(this ClaimsPrincipal principal)
        => (principal?.GetClaimValueAs<long>(JwtClaimTypes.Expiration, ClaimTypes.Expiration)).GetValueOrDefault();

    public static long GetIssuedAt(this ClaimsPrincipal principal)
        => (principal?.GetClaimValueAs<long>(JwtClaimTypes.IssuedAt)).GetValueOrDefault();
}
