using FluentValidation;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Customers.DeleteCustomerAttachment;

internal sealed class DeleteCustomerAttachmentCommandValidator : AbstractValidator<DeleteCustomerAttachmentCommand>
{
    public DeleteCustomerAttachmentCommandValidator()
    {
        RuleFor(x => x.CustomerId)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_ID_REQUIRED);

        RuleFor(x => x.AttachmentId)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_ATTACHMENT_ID_REQUIRED);
    }
}
