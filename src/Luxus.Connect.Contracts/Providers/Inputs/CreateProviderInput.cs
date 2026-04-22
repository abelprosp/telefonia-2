using Luxus.Connect.Contracts.Providers.Commands;

namespace Luxus.Connect.Contracts.Providers.Inputs;

public sealed record CreateProviderInput(
    string Name,
    string Slug
)
{
    public CreateProviderCommand ToCommand(string organizationId)
        => new(organizationId, Name, Slug);
}
