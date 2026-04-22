using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.Customers.Commands;

public sealed record DeleteCustomerAttachmentCommand(string CustomerId, string AttachmentId)
    : ICommand<OneOf<None, AppError>>;
