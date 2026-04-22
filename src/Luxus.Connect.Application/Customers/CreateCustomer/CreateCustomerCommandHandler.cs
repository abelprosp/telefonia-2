using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.Customers.CreateCustomer;

internal sealed class CreateCustomerCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<CreateCustomerCommand, OneOf<CreateCustomerResponse, AppError>>
    , IRequestHandler<CreateCustomerCommand, OneOf<CreateCustomerResponse, AppError>>
{
    public async ValueTask<OneOf<CreateCustomerResponse, AppError>> Handle(CreateCustomerCommand command, CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(new CreateCustomerCommandValidator(), cancellationToken);

        if (!validation.IsValid)
            return new InputValidationError(validation.Errors);

        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        IEnumerable<Customer> sameDocument = await uow.Customers.ListByDocumentAsync(
            appState.Organization!.Id,
            command.Document,
            cancellationToken);

        if (sameDocument.Any())
        {
            return new BusinessRuleError(Notifications.Customers.CUSTOMER_DOCUMENT_DUPLICATED);
        }

        Provider? provider = await uow.Providers.GetByIdAsync(appState.Organization!.Id, command.ProviderId, cancellationToken);

        if (provider is null)
        {
            return new ResourceNotFoundError(Notifications.Providers.PROVIDER_NOT_FOUND);
        }

        var entity = Customer.Create(
           appState.Organization!.Id,
           provider,
           command.Name,
           command.Document);

        if (command.BirthOrOpeningDate.HasValue)
            entity.UpdateBirthOrOpeningDate(command.BirthOrOpeningDate.Value);

        if (!string.IsNullOrWhiteSpace(command.LegalName))
            entity.UpdateLegalName(command.LegalName);

        if (!string.IsNullOrWhiteSpace(command.StateRegistration))
            entity.UpdateStateRegistration(command.StateRegistration);

        if (!string.IsNullOrWhiteSpace(command.ResponsibleSalespersonUserId))
            entity.SetResponsibleSalespersonUserId(command.ResponsibleSalespersonUserId);

        foreach (CreateCustomerAddressCommand address in command.Addresses)
        {
            entity.AddAddress(
                address.Street,
                address.Number,
                address.Neighborhood,
                address.City,
                address.State,
                address.ZipCode,
                address.Complement,
                address.Country);
        }

        await uow.Customers.AddAsync(entity, cancellationToken);

        await uow.CommitAsync(cancellationToken);

        return (CreateCustomerResponse)entity;
    }
}
