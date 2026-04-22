using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.Providers.Commands;

public sealed record InactivateProviderCommand(string OrganizationId, string Id) : ICommand<OneOf<None, AppError>>;
