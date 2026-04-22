using Asp.Versioning;
using Goal.Application.Commands;
using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
using Luxus.Connect.Contracts.Customers.Commands;
using Luxus.Connect.Contracts.Customers.Inputs;
using Luxus.Connect.Contracts.Customers.Responses;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data.Query.Repositories.Customers;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Api.Features.Customers;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class CustomersController(
    ICustomerQueryRepository customerQueryRepository,
    ICustomerProcessingMonthBillingReadinessQueryRepository billingReadinessQueryRepository,
    ICommandSender commandSender,
    AppState appState)
    : ConnectApiController
{
    private const string GET_BY_ID_ROUTE = $"{nameof(CustomersController)}.{nameof(GetById)}";

    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListCustomerResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        [FromQuery(Name = "provider_id")] string? providerId,
        CancellationToken cancellationToken = default)
    {
        IPagedList<ListCustomerResponse> response = await customerQueryRepository.QueryAsync(
            pageSearch.ToPageSearch(),
            providerId,
            cancellationToken);

        return Paged(response);
    }

    [HttpGet("{id}/phone-lines")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<CustomerPhoneLineLinkResponse>>> GetPhoneLines(
        [FromRoute] string id,
        [FromQuery] PageSearchRequest pageSearch,
        CancellationToken cancellationToken = default)
    {
        IPagedList<CustomerPhoneLineLinkResponse> response = await customerQueryRepository.QueryPhoneLinesAsync(
            appState.Organization!.Id,
            id,
            pageSearch.ToPageSearch(),
            cancellationToken);

        return Paged(response);
    }

    [HttpGet("{id}/provider-links")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<IReadOnlyList<CustomerProviderLinkResponse>>> GetProviderLinks(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        IReadOnlyList<CustomerProviderLinkResponse> response = await customerQueryRepository.ListProviderLinksAsync(
            appState.Organization!.Id,
            id,
            cancellationToken);

        return Ok(response);
    }

    [HttpGet("{id}/attachments")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<IReadOnlyList<CustomerAttachmentResponse>>> GetAttachments(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        IReadOnlyList<CustomerAttachmentResponse>? items = await customerQueryRepository.ListAttachmentsAsync(
            appState.Organization!.Id,
            id,
            cancellationToken);

        return items is null
            ? NotFound(ApiResponse.Fail(Notifications.Customers.CUSTOMER_NOT_FOUND))
            : Ok(items);
    }

    [HttpPost("{id}/attachments")]
    [ProducesResponseType(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<CustomerAttachmentResponse>> PostAttachment(
        [FromRoute] string id,
        [FromBody] RegisterCustomerAttachmentInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<CustomerAttachmentResponse, AppError> result = await commandSender.Send(
            input.ToCommand(id),
            cancellationToken);

        return result.Match<ActionResult<CustomerAttachmentResponse>>(
            r => CreatedAtAction(nameof(GetAttachments), new { id }, r),
            err => Error(err));
    }

    [HttpDelete("{id}/attachments/{attachmentId}")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult> DeleteAttachment(
        [FromRoute] string id,
        [FromRoute] string attachmentId,
        CancellationToken cancellationToken = default)
    {
        OneOf<None, AppError> result = await commandSender.Send(
            new DeleteCustomerAttachmentCommand(id, attachmentId),
            cancellationToken);

        return result.Match<ActionResult>(
            _ => Accepted(),
            err => Error(err));
    }

    [HttpGet("{id}/processing-months/{processingMonthId}/billing-readiness")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetCustomerProcessingMonthBillingReadinessResponse>> GetBillingReadiness(
        [FromRoute] string id,
        [FromRoute] string processingMonthId,
        CancellationToken cancellationToken = default)
    {
        GetCustomerProcessingMonthBillingReadinessResponse? response = await billingReadinessQueryRepository.LoadAsync(
            appState.Organization!.Id,
            id,
            processingMonthId,
            cancellationToken);

        return response is null
            ? NotFound(ApiResponse.Fail(Notifications.Customers.CUSTOMER_BILLING_READINESS_CONTEXT_NOT_FOUND))
            : Ok(response);
    }

    [HttpPost("{id}/processing-months/{processingMonthId}/manual-release")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetCustomerProcessingMonthBillingReadinessResponse>> PostManualRelease(
        [FromRoute] string id,
        [FromRoute] string processingMonthId,
        [FromBody] ManuallyReleaseCustomerForProcessingMonthInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<GetCustomerProcessingMonthBillingReadinessResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id, id, processingMonthId),
            cancellationToken);

        return result.Match<ActionResult<GetCustomerProcessingMonthBillingReadinessResponse>>(
            r => Ok(r),
            err => Error(err));
    }

    [HttpGet("{id}", Name = GET_BY_ID_ROUTE)]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<ListCustomerResponse>> GetById([FromRoute] string id, CancellationToken cancellationToken = default)
    {
        ListCustomerResponse? customer = await customerQueryRepository.LoadAsync(id, cancellationToken);

        return customer is null
            ? NotFound(ApiResponse.Fail(Notifications.Customers.CUSTOMER_NOT_FOUND))
            : Ok(customer);
    }

    [HttpPost]
    [ProducesResponseType(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<CreateCustomerResponse>> Post(
        [FromBody] CreateCustomerCommand command,
        CancellationToken cancellationToken = default)
    {
        OneOf<CreateCustomerResponse, AppError> result = await commandSender.Send(command, cancellationToken);

        return result
            .Match<ActionResult<CreateCustomerResponse>>(
                customer => CreatedAtRoute(GET_BY_ID_ROUTE, new { id = customer.Id }, customer),
                error => Error(error)
            );
    }

    [HttpPatch("{id}")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult> Patch([FromRoute] string id, [FromBody] UpdateCustomerInput input)
    {
        OneOf<None, AppError> result = await commandSender.Send(input.ToCommand(id));

        return result
            .Match(
                _ => AcceptedAtRoute(GET_BY_ID_ROUTE, new { id }),
                Error
            );
    }

    [HttpDelete("{id}")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult> Delete([FromRoute] string id)
    {
        OneOf<None, AppError> result = await commandSender.Send(new InactivateCustomerCommand(id));

        return result
            .Match(
                _ => Accepted(),
                Error
            );
    }
}
