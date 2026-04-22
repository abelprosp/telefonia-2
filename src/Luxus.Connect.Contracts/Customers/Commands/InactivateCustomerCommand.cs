using Goal.Application.Commands;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Contracts.Customers.Commands;

public sealed record InactivateCustomerCommand(string Id) : ICommand<OneOf<None, AppError>>;
