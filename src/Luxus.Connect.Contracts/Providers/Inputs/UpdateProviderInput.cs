using Luxus.Connect.Contracts.Providers.Commands;

namespace Luxus.Connect.Contracts.Providers.Inputs;

public sealed record UpdateProviderInput(
    string Name,
    string Slug
)
{
    public UpdateProviderCommand ToCommand(string organizationId, string id)
        => new(organizationId, id, Name, Slug);
}
