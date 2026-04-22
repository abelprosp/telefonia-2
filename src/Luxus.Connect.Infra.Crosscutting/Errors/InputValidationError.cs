using FluentValidation.Results;
using Luxus.Connect.Infra.Crosscutting.Notifications;

namespace Luxus.Connect.Infra.Crosscutting.Errors;

public record InputValidationError(IEnumerable<ValidationFailure> Failures)
    : AppError(ErrorType.InputValidation, Failures.Select(f => new Notification(f.ErrorCode, f.ErrorMessage, f.PropertyName)));