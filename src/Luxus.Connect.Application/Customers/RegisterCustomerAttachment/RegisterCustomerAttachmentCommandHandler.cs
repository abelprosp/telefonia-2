using ConduitR.Abstractions;
using FluentValidation.Results;
using Goal.Application.Commands;
using Goal.Application.Extensions;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data;
using OneOf;

namespace Luxus.Connect.Application.Customers.RegisterCustomerAttachment;

internal sealed class RegisterCustomerAttachmentCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<RegisterCustomerAttachmentCommand, OneOf<CustomerAttachmentResponse, AppError>>
    , IRequestHandler<RegisterCustomerAttachmentCommand, OneOf<CustomerAttachmentResponse, AppError>>
{
    public async ValueTask<OneOf<CustomerAttachmentResponse, AppError>> Handle(
        RegisterCustomerAttachmentCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(
            new RegisterCustomerAttachmentCommandValidator(),
            cancellationToken);

        if (!validation.IsValid)
        {
            return new InputValidationError(validation.Errors);
        }

        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        Customer? customer = await uow.Customers.GetByIdAsync(
            appState.Organization!.Id,
            command.CustomerId,
            cancellationToken);

        if (customer is null)
        {
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_NOT_FOUND);
        }

        CustomerAttachment attachment = CustomerAttachment.Create(
            customer,
            command.Title,
            command.OriginalFileName,
            command.StorageBucket,
            command.StorageObjectKey,
            command.ContentType,
            command.SizeBytes);

        await uow.CustomerAttachments.AddAsync(attachment, cancellationToken);
        await uow.CommitAsync(cancellationToken);

        return ToResponse(attachment);
    }

    private static CustomerAttachmentResponse ToResponse(CustomerAttachment a)
    {
        return new CustomerAttachmentResponse(
            a.Id,
            a.Title,
            a.OriginalFileName,
            a.StorageBucket,
            a.StorageObjectKey,
            a.ContentType,
            a.SizeBytes,
            a.UploadedAtUtc);
    }
}
