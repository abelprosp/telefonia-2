using FluentValidation;
using Luxus.Connect.Contracts.PhoneLines.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.PhoneLines.UnassignPhoneLineCustomer;

internal sealed class UnassignPhoneLineCustomerCommandValidator : AbstractValidator<UnassignPhoneLineCustomerCommand>
{
    public UnassignPhoneLineCustomerCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty()
            .WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.PhoneLineId)
            .NotEmpty()
            .WithNotification(Notifications.PhoneLines.PHONE_LINE_ID_REQUIRED);
    }
}
