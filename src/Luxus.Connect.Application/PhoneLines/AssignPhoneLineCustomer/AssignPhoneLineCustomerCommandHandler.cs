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

namespace Luxus.Connect.Application.PhoneLines.AssignPhoneLineCustomer;

internal sealed class AssignPhoneLineCustomerCommandHandler(IAppUnitOfWork uow)
    : ICommandHandler<AssignPhoneLineCustomerCommand, OneOf<PhoneLineCustomerLinkResponse, AppError>>,
      IRequestHandler<AssignPhoneLineCustomerCommand, OneOf<PhoneLineCustomerLinkResponse, AppError>>
{
    public async ValueTask<OneOf<PhoneLineCustomerLinkResponse, AppError>> Handle(
        AssignPhoneLineCustomerCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation =
            await command.ValidateCommandAsync(new AssignPhoneLineCustomerCommandValidator(), cancellationToken);

        if (!validation.IsValid)
        {
            return new InputValidationError(validation.Errors);
        }

        PhoneLine? line = await uow.PhoneLines.GetByIdAsync(command.OrganizationId, command.PhoneLineId, cancellationToken);
        if (line is null)
        {
            return new ResourceNotFoundError(Notifications.PhoneLines.PHONE_LINE_NOT_FOUND);
        }

        Customer? customer = await uow.Customers.GetByIdAsync(command.OrganizationId, command.CustomerId, cancellationToken);
        if (customer is null)
        {
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_NOT_FOUND);
        }

        string? previousCustomerId = line.ActiveCustomerLink?.CustomerId;
        line.AssignCustomer(customer, command.StartDate);
        customer.AddProviderLink(line.ProviderAccount.ContractingCompany.Provider, command.StartDate);
        customer.Reactivate();

        if (!string.IsNullOrWhiteSpace(previousCustomerId) && previousCustomerId != customer.Id)
        {
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
        }

        await uow.CommitAsync(cancellationToken);

        PhoneLineCustomerLink activeLink = line.ActiveCustomerLink!;
        return new PhoneLineCustomerLinkResponse(
            line.Id,
            activeLink.CustomerId,
            activeLink.Customer.Name,
            activeLink.Customer.Documents
                .FirstOrDefault(d => d.DocumentType is CustomerDocumentType.CPF or CustomerDocumentType.CNPJ)
                ?.Number,
            activeLink.StartDate,
            activeLink.EndDate,
            activeLink.IsActive);
    }
}
