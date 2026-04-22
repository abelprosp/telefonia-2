using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.Providers.CreateProvider;

public sealed class CreateProviderCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<CreateProviderCommand, OneOf<CreateProviderResponse, AppError>>
    , IRequestHandler<CreateProviderCommand, OneOf<CreateProviderResponse, AppError>>
{
    public async ValueTask<OneOf<CreateProviderResponse, AppError>> Handle(CreateProviderCommand command, CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new CreateProviderCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        Provider? existing = await uow.Providers.GetBySlugAsync(
            command.OrganizationId,
            command.Slug,
            cancellationToken);

        if (existing is not null)
        {
            return new BusinessRuleError(Notifications.Providers.PROVIDER_SLUG_DUPLICATED);
        }

        var entity = Provider.Create(
            command.OrganizationId,
            command.Name,
            command.Slug);

        await uow.Providers.AddAsync(entity, cancellationToken);

        await uow.CommitAsync(cancellationToken);

        return (CreateProviderResponse)entity;
    }
}
