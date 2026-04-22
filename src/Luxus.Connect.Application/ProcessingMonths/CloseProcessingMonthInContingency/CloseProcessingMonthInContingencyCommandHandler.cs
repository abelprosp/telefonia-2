using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.ProcessingMonths.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.ProcessingMonths.CloseProcessingMonthInContingency;

internal sealed class CloseProcessingMonthInContingencyCommandHandler(IAppUnitOfWork uow, AppState appState)
    : ICommandHandler<CloseProcessingMonthInContingencyCommand, OneOf<GetProcessingMonthResponse, AppError>>,
      IRequestHandler<CloseProcessingMonthInContingencyCommand, OneOf<GetProcessingMonthResponse, AppError>>
{
    public async ValueTask<OneOf<GetProcessingMonthResponse, AppError>> Handle(
        CloseProcessingMonthInContingencyCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new CloseProcessingMonthInContingencyCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);

        ProcessingMonth? entity = await uow.ProcessingMonths.GetByIdAsync(command.OrganizationId, command.Id, cancellationToken);

        if (entity is null)
            return new ResourceNotFoundError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_FOUND);

        if (entity.Status == ProcessingMonthStatus.CLOSED)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_ALREADY_CLOSED);

        entity.CloseInContingency(appState.User.UserId, command.Justification);

        await uow.CommitAsync(cancellationToken);

        return (GetProcessingMonthResponse)entity;
    }
}
