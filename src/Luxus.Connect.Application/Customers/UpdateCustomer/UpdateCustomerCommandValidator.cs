using FluentValidation;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Customers.UpdateCustomer;

internal sealed class UpdateCustomerCommandValidator : AbstractValidator<UpdateCustomerCommand>
{
    public UpdateCustomerCommandValidator()
    {
        RuleFor(x => x.Id).NotEmpty();
        RuleFor(x => x.Name)
            .NotEmpty().WithNotification(Notifications.Customers.CUSTOMER_NAME_REQUIRED)
            .MaximumLength(256).WithNotification(Notifications.Customers.CUSTOMER_NAME_MAX_LENGTH);

        RuleFor(x => x.ResponsibleSalespersonUserId)
            .MaximumLength(256)
            .When(x => !string.IsNullOrWhiteSpace(x.ResponsibleSalespersonUserId))
            .WithNotification(Notifications.Customers.CUSTOMER_RESPONSIBLE_SALESPERSON_USER_ID_MAX_LENGTH);
    }
}
