using FluentValidation;
using Luxus.Connect.Contracts.PhoneLines.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.PhoneLines.AssignPhoneLineCustomer;

internal sealed class AssignPhoneLineCustomerCommandValidator : AbstractValidator<AssignPhoneLineCustomerCommand>
{
    public AssignPhoneLineCustomerCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty()
            .WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.PhoneLineId)
            .NotEmpty()
            .WithNotification(Notifications.PhoneLines.PHONE_LINE_ID_REQUIRED);

        RuleFor(x => x.CustomerId)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_ID_REQUIRED);
    }
}
