using Goal.Application.Commands;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Contracts.Customers.Commands;

public sealed record RegisterCustomerAttachmentCommand : ICommand<OneOf<CustomerAttachmentResponse, AppError>>
{
    public required string CustomerId { get; init; }
    public string? Title { get; init; }
    public required string OriginalFileName { get; init; }
    public required string StorageBucket { get; init; }
    public required string StorageObjectKey { get; init; }
    public string? ContentType { get; init; }
    public long? SizeBytes { get; init; }
}
