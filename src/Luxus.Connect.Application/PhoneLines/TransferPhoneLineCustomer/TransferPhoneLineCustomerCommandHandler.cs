using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.PhoneLines.Commands;
using Luxus.Connect.Contracts.PhoneLines.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.PhoneLines.TransferPhoneLineCustomer;

internal sealed class TransferPhoneLineCustomerCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<TransferPhoneLineCustomerCommand, OneOf<PhoneLineCustomerLinkResponse, AppError>>,
      IRequestHandler<TransferPhoneLineCustomerCommand, OneOf<PhoneLineCustomerLinkResponse, AppError>>
{
    public async ValueTask<OneOf<PhoneLineCustomerLinkResponse, AppError>> Handle(
        TransferPhoneLineCustomerCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation =
            await command.ValidateCommandAsync(new TransferPhoneLineCustomerCommandValidator(), cancellationToken);

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

        Customer? customer = await uow.Customers.GetByIdAsync(command.OrganizationId, command.CustomerId, cancellationToken);
        if (customer is null)
        {
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_NOT_FOUND);
        }

        if (line.ActiveCustomerLink.CustomerId == customer.Id)
        {
            return new BusinessRuleError(Notifications.PhoneLines.PHONE_LINE_CUSTOMER_TRANSFER_SAME_CUSTOMER);
        }

        string previousCustomerId = line.ActiveCustomerLink.CustomerId;
        line.AssignCustomer(customer, command.TransferDate);
        customer.AddProviderLink(line.ProviderAccount.ContractingCompany.Provider, command.TransferDate);
        customer.Reactivate();

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

        PhoneLineCustomerLink activeLink = line.ActiveCustomerLink!;

        return new PhoneLineCustomerLinkResponse(
            line.Id,
            activeLink.CustomerId,
            activeLink.Customer.Name,
            activeLink.Customer.GetCpfOrCnpj(),
            activeLink.StartDate,
            activeLink.EndDate,
            activeLink.IsActive);
    }
}
