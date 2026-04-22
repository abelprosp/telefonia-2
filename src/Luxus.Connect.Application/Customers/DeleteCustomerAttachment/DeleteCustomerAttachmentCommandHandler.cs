using ConduitR.Abstractions;
using FluentValidation.Results;
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

namespace Luxus.Connect.Application.Customers.DeleteCustomerAttachment;

internal sealed class DeleteCustomerAttachmentCommandHandler(
    IAppUnitOfWork uow,
    AppState appState)
    : ICommandHandler<DeleteCustomerAttachmentCommand, OneOf<None, AppError>>
    , IRequestHandler<DeleteCustomerAttachmentCommand, OneOf<None, AppError>>
{
    public async ValueTask<OneOf<None, AppError>> Handle(
        DeleteCustomerAttachmentCommand command,
        CancellationToken cancellationToken)
    {
        ValidationResult validation = await command.ValidateCommandAsync(
            new DeleteCustomerAttachmentCommandValidator(),
            cancellationToken);

        if (!validation.IsValid)
        {
            return new InputValidationError(validation.Errors);
        }

        if (appState.User is null)
        {
            return new BusinessRuleError(Notifications.Shared.DOMAIN_VIOLATION);
        }

        CustomerAttachment? attachment = await uow.CustomerAttachments.GetByIdAsync(
            appState.Organization!.Id,
            command.CustomerId,
            command.AttachmentId,
            cancellationToken);

        if (attachment is null)
        {
            return new ResourceNotFoundError(Notifications.Customers.CUSTOMER_ATTACHMENT_NOT_FOUND);
        }

        uow.CustomerAttachments.Remove(attachment);
        await uow.CommitAsync(cancellationToken);

        return default(None);
    }
}
