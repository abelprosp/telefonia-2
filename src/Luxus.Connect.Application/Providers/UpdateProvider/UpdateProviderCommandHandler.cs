using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Application.Providers.UpdateProvider;

internal sealed class UpdateProviderCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<UpdateProviderCommand, OneOf<None, AppError>>
    , IRequestHandler<UpdateProviderCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(UpdateProviderCommand command, CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new UpdateProviderCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        Provider? entity = await uow.Providers.GetByIdAsync(
            command.OrganizationId,
            command.Id,
            cancellationToken);

        if (entity is null)
        {
            return new ResourceNotFoundError(Notifications.Providers.PROVIDER_NOT_FOUND);
        }

        Provider? existingBySlug = await uow.Providers.GetBySlugAsync(
            command.OrganizationId,
            command.Slug,
            cancellationToken);

        if (existingBySlug is not null && existingBySlug.Id != command.Id)
        {
            return new BusinessRuleError(Notifications.Providers.PROVIDER_SLUG_DUPLICATED);
        }

        entity.Update(command.Name, command.Slug);

        uow.Providers.Update(entity);
        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
