using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.PhoneLines.Commands;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Application.PhoneLines.UnassignPhoneLineCustomer;

internal sealed class UnassignPhoneLineCustomerCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<UnassignPhoneLineCustomerCommand, OneOf<None, AppError>>,
      IRequestHandler<UnassignPhoneLineCustomerCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(
        UnassignPhoneLineCustomerCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation =
            await command.ValidateCommandAsync(new UnassignPhoneLineCustomerCommandValidator(), cancellationToken);

        if (!validation.IsValid)
        {
            return new InputValidationError(validation.Errors);
        }

        PhoneLine? line = await uow.PhoneLines.GetByIdAsync(command.OrganizationId, command.PhoneLineId, cancellationToken);
        if (line is null)
        {
            return new ResourceNotFoundError(Notifications.PhoneLines.PHONE_LINE_NOT_FOUND);
        }

        if (line.ActiveCustomerLink is null)
        {
            return new BusinessRuleError(Notifications.PhoneLines.PHONE_LINE_ACTIVE_CUSTOMER_LINK_NOT_FOUND);
        }

        string previousCustomerId = line.ActiveCustomerLink.CustomerId;
        line.UnassignCustomer(command.EndDate);

        Customer? previousCustomer = await uow.Customers.GetByIdAsync(
            command.OrganizationId,
            previousCustomerId,
            cancellationToken);

        if (previousCustomer is not null)
        {
            bool previousStillHasActiveLines = await uow.Customers.HasActivePhoneLinesAsync(
                command.OrganizationId,
                previousCustomer.Id,
                line.Id,
                cancellationToken);

            if (!previousStillHasActiveLines)
            {
                previousCustomer.Inactivate();
            }
        }

        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
