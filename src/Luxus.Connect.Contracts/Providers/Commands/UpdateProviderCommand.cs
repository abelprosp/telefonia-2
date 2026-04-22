using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.Providers.Commands;

public sealed record UpdateProviderCommand(
    string OrganizationId,
    string Id,
    string Name,
    string Slug
) : ICommand<OneOf<None, AppError>>;
