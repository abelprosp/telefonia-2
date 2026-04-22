using FluentValidation;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Customers.CreateCustomer;

internal sealed class CreateCustomerCommandValidator : AbstractValidator<CreateCustomerCommand>
{
    public CreateCustomerCommandValidator()
    {
        RuleFor(x => x.ProviderId)
            .NotEmpty().WithNotification(Notifications.Customers.CUSTOMER_PROVIDER_REQUIRED);

        RuleFor(x => x.Name)
            .NotEmpty().WithNotification(Notifications.Customers.CUSTOMER_NAME_REQUIRED)
            .MaximumLength(256).WithNotification(Notifications.Customers.CUSTOMER_NAME_MAX_LENGTH);

        RuleFor(x => x.Document)
            .NotEmpty().WithNotification(Notifications.Customers.CUSTOMER_DOCUMENT_REQUIRED)
            .MaximumLength(20).WithNotification(Notifications.Customers.CUSTOMER_DOCUMENT_MAX_LENGTH);

        RuleFor(x => x.ResponsibleSalespersonUserId)
            .MaximumLength(256)
            .When(x => !string.IsNullOrWhiteSpace(x.ResponsibleSalespersonUserId))
            .WithNotification(Notifications.Customers.CUSTOMER_RESPONSIBLE_SALESPERSON_USER_ID_MAX_LENGTH);

        RuleFor(x => x.LegalName)
            .NotEmpty()
            .When(x => x.Type == "PJ")
            .WithNotification(Notifications.Customers.CUSTOMER_LEGAL_NAME_REQUIRED_FOR_PJ);

        RuleForEach(x => x.Addresses).SetValidator(new CreateCustomerAddressInputValidator());
    }
}

internal sealed class CreateCustomerAddressInputValidator : AbstractValidator<CreateCustomerAddressCommand>
{
    public CreateCustomerAddressInputValidator()
    {
        RuleFor(x => x.Street).NotEmpty().MaximumLength(256);
        RuleFor(x => x.Neighborhood).NotEmpty().MaximumLength(256);
        RuleFor(x => x.Number).NotEmpty().MaximumLength(20);
        RuleFor(x => x.City).NotEmpty().MaximumLength(100);
        RuleFor(x => x.State).NotEmpty().MaximumLength(2);
        RuleFor(x => x.ZipCode).NotEmpty().MaximumLength(10);
    }
}
