using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.ProcessingMonths.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.ProcessingMonths.CreateProcessingMonth;

internal sealed class CreateProcessingMonthCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<CreateProcessingMonthCommand, OneOf<GetProcessingMonthResponse, AppError>>,
      IRequestHandler<CreateProcessingMonthCommand, OneOf<GetProcessingMonthResponse, AppError>>
{
    public async ValueTask<OneOf<GetProcessingMonthResponse, AppError>> Handle(
        CreateProcessingMonthCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new CreateProcessingMonthCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        Provider? provider = await uow.Providers.GetByIdAsync(command.OrganizationId, command.ProviderId, cancellationToken);

        if (provider is null)
            return new ResourceNotFoundError(Notifications.Providers.PROVIDER_NOT_FOUND);

        ProcessingMonth? duplicated = await uow.ProcessingMonths.GetByProviderAndCalendarAsync(
            provider.Id,
            command.Year,
            command.Month,
            cancellationToken);

        if (duplicated is not null)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_DUPLICATE);

        var entity = ProcessingMonth.Create(provider, command.Year, command.Month, command.DisplayName);

        await uow.ProcessingMonths.AddAsync(entity, cancellationToken);
        await uow.CommitAsync(cancellationToken);

        return (GetProcessingMonthResponse)entity;
    }
}
