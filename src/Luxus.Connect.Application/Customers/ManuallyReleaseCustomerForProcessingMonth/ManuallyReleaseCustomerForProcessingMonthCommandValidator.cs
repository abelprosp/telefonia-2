using FluentValidation;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Customers.ManuallyReleaseCustomerForProcessingMonth;

internal sealed class ManuallyReleaseCustomerForProcessingMonthCommandValidator
    : AbstractValidator<ManuallyReleaseCustomerForProcessingMonthCommand>
{
    public ManuallyReleaseCustomerForProcessingMonthCommandValidator()
    {
        RuleFor(x => x.OrganizationId)
            .NotEmpty()
            .WithNotification(Notifications.Shared.ORGANIZATION_ID_REQUIRED);

        RuleFor(x => x.CustomerId)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_ID_REQUIRED);

        RuleFor(x => x.ProcessingMonthId)
            .NotEmpty()
            .WithNotification(Notifications.Invoices.INVOICE_PROCESSING_MONTH_REQUIRED);

        RuleFor(x => x.Justification)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_REQUIRED)
            .MinimumLength(10)
            .WithNotification(Notifications.Customers.CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_REQUIRED)
            .MaximumLength(4000)
            .WithNotification(Notifications.Customers.CUSTOMER_MANUAL_RELEASE_JUSTIFICATION_MAX_LENGTH);
    }
}
