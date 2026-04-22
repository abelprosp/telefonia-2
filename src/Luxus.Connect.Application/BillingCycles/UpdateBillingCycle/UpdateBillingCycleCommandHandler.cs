using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.BillingCycles.Commands;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Application.BillingCycles.UpdateBillingCycle;

internal sealed class UpdateBillingCycleCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<UpdateBillingCycleCommand, OneOf<None, AppError>>
    , IRequestHandler<UpdateBillingCycleCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(UpdateBillingCycleCommand command, CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new UpdateBillingCycleCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        Provider? provider = await uow.Providers.GetByIdAsync(
            command.OrganizationId,
            command.ProviderId,
            cancellationToken);

        if (provider is null)
        {
            return new BusinessRuleError(Notifications.Providers.PROVIDER_NOT_FOUND);
        }

        BillingCycle? entity = await uow.BillingCycles.GetByIdAsync(command.OrganizationId, command.Id, cancellationToken);

        if (entity is null)
        {
            return new ResourceNotFoundError(Notifications.BillingCycles.BILLING_CYCLE_NOT_FOUND);
        }

        if (entity.Status == BillingCycleStatus.CLOSED)
        {
            return new BusinessRuleError(Notifications.BillingCycles.BILLING_CYCLE_CONSOLIDATED);
        }

        if (await uow.ProcessingMonths.ExistsClosedIntersectingDateRangeAsync(
            command.OrganizationId,
            entity.ProviderId,
            entity.StartDate,
            entity.EndDate,
            cancellationToken))
        {
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED);
        }

        if (await uow.ProcessingMonths.ExistsClosedIntersectingDateRangeAsync(
            command.OrganizationId,
            entity.ProviderId,
            command.StartDate,
            command.EndDate,
            cancellationToken))
        {
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED);
        }

        entity.Update(command.Code, command.Name, command.StartDate, command.EndDate);
        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
