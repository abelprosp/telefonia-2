using Asp.Versioning;
using Goal.Application.Commands;
using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
using Luxus.Connect.Contracts.PhoneLines.Inputs;
using Luxus.Connect.Contracts.PhoneLines.Responses;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data.Query.Repositories.PhoneLines;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Api.Features.PhoneLines;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class PhoneLinesController(
    IPhoneLineQueryRepository phoneLineQueryRepository,
    ICommandSender commandSender,
    AppState appState)
    : ConnectApiController
{
    private const string GET_BY_ID_ROUTE = $"{nameof(PhoneLinesController)}.{nameof(GetById)}";

    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListPhoneLineResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        [FromQuery] string? status,
        CancellationToken cancellationToken = default)
    {
        IPagedList<ListPhoneLineResponse> response = string.IsNullOrWhiteSpace(status)
            ? await phoneLineQueryRepository.QueryAsync(appState.Organization!.Id, pageSearch.ToPageSearch(), cancellationToken)
            : await phoneLineQueryRepository.QueryByStatusAsync(appState.Organization!.Id, status, pageSearch.ToPageSearch(), cancellationToken);

        return Paged(response);
    }

    [HttpGet("{id}", Name = GET_BY_ID_ROUTE)]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetPhoneLineResponse>> GetById(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        GetPhoneLineResponse? line = await phoneLineQueryRepository.LoadAsync(appState.Organization!.Id, id, cancellationToken);

        return line is null
            ? NotFound(ApiResponse.Fail(Notifications.PhoneLines.PHONE_LINE_NOT_FOUND))
            : Ok(line);
    }

    [HttpGet("{id}/customer-links")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<IEnumerable<PhoneLineCustomerLinkResponse>>> GetCustomerLinks(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        IEnumerable<PhoneLineCustomerLinkResponse> links = await phoneLineQueryRepository.ListCustomerLinksAsync(
            appState.Organization!.Id,
            id,
            cancellationToken);

        return Ok(links);
    }

    [HttpPost("{id}/customer-links")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PhoneLineCustomerLinkResponse>> AssignCustomer(
        [FromRoute] string id,
        [FromBody] AssignPhoneLineCustomerInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<PhoneLineCustomerLinkResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id, id),
            cancellationToken);

        return result.Match<ActionResult<PhoneLineCustomerLinkResponse>>(
            r => Ok(r),
            err => Error(err));
    }

    [HttpPost("{id}/customer-links/transfer")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PhoneLineCustomerLinkResponse>> TransferCustomer(
        [FromRoute] string id,
        [FromBody] TransferPhoneLineCustomerInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<PhoneLineCustomerLinkResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id, id),
            cancellationToken);

        return result.Match<ActionResult<PhoneLineCustomerLinkResponse>>(
            r => Ok(r),
            err => Error(err));
    }

    [HttpDelete("{id}/customer-links/active")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult> UnassignCustomer(
        [FromRoute] string id,
        [FromBody] UnassignPhoneLineCustomerInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<None, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id, id),
            cancellationToken);

        return result.Match<ActionResult>(
            _ => Accepted(),
            Error);
    }
}
