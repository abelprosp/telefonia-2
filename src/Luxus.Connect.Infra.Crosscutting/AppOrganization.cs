using System.Security.Claims;
using System.Text.Json.Serialization;
using Luxus.Connect.Infra.Crosscutting.Extensions;

namespace Luxus.Connect.Infra.Crosscutting;

public sealed class AppOrganization
{
    public AppOrganization(ClaimsPrincipal principal)
    {
        Dictionary<string, OrganizationClaim> orgs = principal.GetOrganization<OrganizationClaim>()!;

        KeyValuePair<string, OrganizationClaim> org = orgs.First();

        Id = org.Value.Id;
        Name = org.Value.Names[0];
        Alias = org.Key;
    }

    public string Id { get; init; } = default!;
    public string Name { get; init; } = default!;
    public string Alias { get; init; } = default!;
}

internal class OrganizationClaim
{
    [JsonPropertyName("name")]
    public string[] Names { get; set; } = [];

    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;
}