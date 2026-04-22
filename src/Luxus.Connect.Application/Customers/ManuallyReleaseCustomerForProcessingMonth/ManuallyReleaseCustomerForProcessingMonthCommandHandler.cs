using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using Luxus.Connect.Infra.Data.Query.Repositories.Customers;
using OneOf;

namespace Luxus.Connect.Application.Customers.ManuallyReleaseCustomerForProcessingMonth;

internal sealed class ManuallyReleaseCustomerForProcessingMonthCommandHandler(
    IAppUnitOfWork uow,
    AppState appState,
    ICustomerProcessingMonthBillingReadinessQueryRepository billingReadinessQuery)
    : ICommandHandler<ManuallyReleaseCustomerForProcessingMonthCommand, OneOf<GetCustomerProcessingMonthBillingReadinessResponse, AppError>>,
      IRequestHandler<ManuallyReleaseCustomerForProcessingMonthCommand, OneOf<GetCustomerProcessingMonthBillingReadinessResponse, AppError>>
{
    public async ValueTask<OneOf<GetCustomerProcessingMonthBillingReadinessResponse, AppError>> Handle(
        ManuallyReleaseCustomerForProcessingMonthCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(
            new ManuallyReleaseCustomerForProcessingMonthCommandValidator(),
            cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);

        Customer? customer = await uow.Customers.GetByIdAsync(command.OrganizationId, command.CustomerId, cancellationToken);

        if (customer is null)
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_NOT_FOUND);

        ProcessingMonth? month =
            await uow.ProcessingMonths.GetByIdAsync(command.OrganizationId, command.ProcessingMonthId, cancellationToken);

        if (month is null)
            return new ResourceNotFoundError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_FOUND);

        if (!customer.HasActiveProvider(month.ProviderId))
            return new BusinessRuleError(Notifications.Customers.CUSTOMER_PROCESSING_MONTH_PROVIDER_MISMATCH);

        if (month.Status != ProcessingMonthStatus.OPEN)
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_OPEN);

        CustomerProcessingMonthManualRelease? existing =
            await uow.CustomerProcessingMonthManualReleases.GetAsync(
                command.OrganizationId,
                command.CustomerId,
                command.ProcessingMonthId,
                cancellationToken);

        if (existing is not null)
            return new BusinessRuleError(Notifications.Customers.CUSTOMER_MANUAL_RELEASE_ALREADY_EXISTS);

        var entity = CustomerProcessingMonthManualRelease.Create(
            customer,
            month,
            command.Justification,
            appState.User.UserId);

        await uow.CustomerProcessingMonthManualReleases.AddAsync(entity, cancellationToken);
        await uow.CommitAsync(cancellationToken);

        GetCustomerProcessingMonthBillingReadinessResponse? response = await billingReadinessQuery.LoadAsync(
            command.OrganizationId,
            command.CustomerId,
            command.ProcessingMonthId,
            cancellationToken);

        if (response is null)
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);

        return response;
    }
}
