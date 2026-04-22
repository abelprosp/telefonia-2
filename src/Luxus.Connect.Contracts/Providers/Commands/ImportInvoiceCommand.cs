using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.Providers.Commands;

public sealed record ImportInvoiceCommand(string ImportRequestId)
    : ICommand<OneOf<None, AppError>>;