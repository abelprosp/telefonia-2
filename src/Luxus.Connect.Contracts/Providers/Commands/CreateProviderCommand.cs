using Goal.Application.Commands;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.Providers.Commands;

public sealed record CreateProviderCommand(
    string OrganizationId,
    string Name,
    string Slug
) : ICommand<OneOf<CreateProviderResponse, AppError>>;
