using FluentValidation;
using Luxus.Connect.Contracts.BillingCycles.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.BillingCycles.UpdateBillingCycle;

internal sealed class UpdateBillingCycleCommandValidator : AbstractValidator<UpdateBillingCycleCommand>
{
    public UpdateBillingCycleCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty().WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED)
            .MaximumLength(100).WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.Code)
            .NotEmpty()
            .WithNotification(Notifications.BillingCycles.BILLING_CYCLE_CODE_REQUIRED);

        RuleFor(x => x.Name)
            .NotEmpty()
            .WithNotification(Notifications.BillingCycles.BILLING_CYCLE_NAME_REQUIRED);
    }
}
