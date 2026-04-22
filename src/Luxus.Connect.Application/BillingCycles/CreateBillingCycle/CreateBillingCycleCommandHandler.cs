using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.BillingCycles.Commands;
using Luxus.Connect.Contracts.BillingCycles.Responses;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.BillingCycles.CreateBillingCycle;

internal sealed class CreateBillingCycleCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<CreateBillingCycleCommand, OneOf<CreateBillingCycleResponse, AppError>>
    , IRequestHandler<CreateBillingCycleCommand, OneOf<CreateBillingCycleResponse, AppError>>
{
    public async ValueTask<OneOf<CreateBillingCycleResponse, AppError>> Handle(CreateBillingCycleCommand command, CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new CreateBillingCycleCommandValidator(), cancellationToken);

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

        if (provider.OrganizationId != command.OrganizationId)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        if (await uow.ProcessingMonths.ExistsClosedIntersectingDateRangeAsync(
                command.OrganizationId,
                provider.Id,
                command.StartDate,
                command.EndDate,
                cancellationToken))
        {
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED);
        }

        var entity = BillingCycle.Create(
            provider,
            command.Code,
            command.Name,
            command.StartDate,
            command.EndDate);

        await uow.BillingCycles.AddAsync(entity, cancellationToken);
        await uow.CommitAsync(cancellationToken);

        return (CreateBillingCycleResponse)entity;
    }
}
