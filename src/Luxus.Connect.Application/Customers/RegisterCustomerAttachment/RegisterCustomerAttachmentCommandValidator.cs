using FluentValidation;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.ObjectStorage;
using Luxus.Connect.Infra.Crosscutting.Validations.Fluent;

namespace Luxus.Connect.Application.Customers.RegisterCustomerAttachment;

internal sealed class RegisterCustomerAttachmentCommandValidator : AbstractValidator<RegisterCustomerAttachmentCommand>
{
    private const long MaxSizeBytes = 256L * 1024 * 1024;

    public RegisterCustomerAttachmentCommandValidator()
    {
        RuleFor(x => x.CustomerId)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_ID_REQUIRED);

        RuleFor(x => x.Title)
            .MaximumLength(256)
            .When(x => x.Title is not null)
            .WithNotification(Notifications.Customers.CUSTOMER_ATTACHMENT_TITLE_MAX_LENGTH);

        RuleFor(x => x.OriginalFileName)
            .NotEmpty()
            .WithNotification(Notifications.Customers.CUSTOMER_ATTACHMENT_ORIGINAL_FILE_NAME_REQUIRED)
            .MaximumLength(512)
            .WithNotification(Notifications.InvoiceImports.ORIGINAL_FILE_NAME_MAX_LENGTH);

        RuleFor(x => x.StorageBucket)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.STORAGE_BUCKET_REQUIRED)
            .MaximumLength(256)
            .WithNotification(Notifications.InvoiceImports.STORAGE_BUCKET_MAX_LENGTH);

        RuleFor(x => x.StorageObjectKey)
            .NotEmpty()
            .WithNotification(Notifications.InvoiceImports.STORAGE_OBJECT_KEY_REQUIRED)
            .MaximumLength(2048)
            .WithNotification(Notifications.InvoiceImports.STORAGE_OBJECT_KEY_MAX_LENGTH)
            .Must(k => !ObjectStorageObjectKeyRules.IsInvalid(k))
            .WithNotification(Notifications.ObjectStorage.OBJECT_KEY_INVALID);

        RuleFor(x => x.ContentType)
            .MaximumLength(128)
            .When(x => x.ContentType is not null)
            .WithNotification(Notifications.Customers.CUSTOMER_ATTACHMENT_CONTENT_TYPE_MAX_LENGTH);

        RuleFor(x => x.SizeBytes)
            .LessThanOrEqualTo(MaxSizeBytes)
            .When(x => x.SizeBytes is not null)
            .WithNotification(Notifications.Customers.CUSTOMER_ATTACHMENT_SIZE_BYTES_TOO_LARGE);
    }
}
