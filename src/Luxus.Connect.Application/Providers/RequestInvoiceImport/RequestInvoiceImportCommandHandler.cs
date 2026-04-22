using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Contracts.Providers.Events;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Bus;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.Providers.RequestInvoiceImport;

internal sealed class RequestInvoiceImportCommandHandler(
    IAppUnitOfWork uow,
    IBusPublisher busPublisher,
    AppState appState)
    : ICommandHandler<ProviderInvoiceImportRequestCommand, OneOf<RequestProviderInvoiceImportResponse, AppError>>, IRequestHandler<ProviderInvoiceImportRequestCommand, OneOf<RequestProviderInvoiceImportResponse, AppError>>
{
    public async ValueTask<OneOf<RequestProviderInvoiceImportResponse, AppError>> Handle(
        ProviderInvoiceImportRequestCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation =
            await command.ValidateCommandAsync(new RequestInvoiceImportCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);

        Provider? provider = await uow.Providers.GetByIdAsync(command.OrganizationId, command.ProviderId, cancellationToken);

        if (provider is null)
            return new BusinessRuleError(Notifications.Providers.PROVIDER_NOT_FOUND);

        ProcessingMonth? processingMonth =
            await uow.ProcessingMonths.GetByIdAsync(command.OrganizationId, command.ProcessingMonthId, cancellationToken);

        if (processingMonth is null)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_FOUND);

        if (processingMonth.ProviderId != command.ProviderId)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_PROVIDER_MISMATCH);

        if (processingMonth.Status != ProcessingMonthStatus.OPEN)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_OPEN);

        var entity = ProviderInvoiceImportRequest.Create(
            command.ProviderId,
            command.ProcessingMonthId,
            command.StorageBucket,
            command.StorageObjectKey,
            command.OrganizationId,
            appState.User.UserId);

        entity.SetOriginalFileName(command.OriginalFileName);

        await uow.InvoiceImportRequests.AddAsync(entity, cancellationToken);

        await uow.CommitAsync(cancellationToken);

        var @event = new InvoiceImportRequestedEvent(
            entity.Id,
            entity.StorageBucket,
            entity.StorageObjectKey,
            entity.OriginalFileName,
            entity.CreatedBy!);

        await busPublisher.Publish(
            @event,
            cancellationToken);

        return (RequestProviderInvoiceImportResponse)entity;
    }
}
