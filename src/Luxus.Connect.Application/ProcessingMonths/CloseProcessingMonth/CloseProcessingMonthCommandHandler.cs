using ConduitR.Abstractions;
using Goal.Application.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.ProcessingMonths.CloseProcessingMonth;

internal sealed class CloseProcessingMonthCommandHandler(IAppUnitOfWork uow, AppState appState)
    : ICommandHandler<CloseProcessingMonthCommand, OneOf<GetProcessingMonthResponse, AppError>>,
      IRequestHandler<CloseProcessingMonthCommand, OneOf<GetProcessingMonthResponse, AppError>>
{
    public async ValueTask<OneOf<GetProcessingMonthResponse, AppError>> Handle(
        CloseProcessingMonthCommand command,
        CancellationToken cancellationToken)
    {
        if (appState.User is null)
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);

        ProcessingMonth? entity = await uow.ProcessingMonths.GetByIdAsync(command.OrganizationId, command.Id, cancellationToken);

        if (entity is null)
            return new ResourceNotFoundError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_FOUND);

        if (entity.Status == ProcessingMonthStatus.CLOSED)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_ALREADY_CLOSED);

        entity.Close(appState.User.UserId);

        await uow.CommitAsync(cancellationToken);

        return (GetProcessingMonthResponse)entity;
    }
}
