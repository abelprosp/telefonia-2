using ConduitR.Abstractions;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Application.Customers.UpdateCustomer;

internal sealed class UpdateCustomerCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<UpdateCustomerCommand, OneOf<None, AppError>>
    , IRequestHandler<UpdateCustomerCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(UpdateCustomerCommand command, CancellationToken cancellationToken)
    {
        FluentValidation.Results.ValidationResult validation = await command.ValidateCommandAsync(new UpdateCustomerCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        Customer? entity = await uow.Customers.GetByIdAsync(
            appState.Organization!.Id,
            command.Id,
            cancellationToken);

        if (entity is null)
        {
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_NOT_FOUND);
        }

        if (entity.Type == CustomerType.PJ && string.IsNullOrWhiteSpace(command.LegalName))
        {
            return new BusinessRuleError(Notifications.Customers.CUSTOMER_LEGAL_NAME_REQUIRED_FOR_PJ);
        }

        if (!string.IsNullOrWhiteSpace(command.Name))
        {
            entity.UpdateName(command.Name);
        }

        if (!string.IsNullOrWhiteSpace(command.LegalName))
        {
            entity.UpdateLegalName(command.LegalName);
        }

        if (!string.IsNullOrWhiteSpace(command.StateRegistration))
        {
            entity.UpdateStateRegistration(command.StateRegistration);
        }

        entity.SetResponsibleSalespersonUserId(command.ResponsibleSalespersonUserId);

        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
