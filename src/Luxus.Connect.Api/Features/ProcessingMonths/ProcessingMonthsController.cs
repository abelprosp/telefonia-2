using Asp.Versioning;
using Goal.Application.Commands;
using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
using Luxus.Connect.Contracts.ProcessingMonths.Commands;
using Luxus.Connect.Contracts.ProcessingMonths.Inputs;
using Luxus.Connect.Contracts.ProcessingMonths.Responses;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data.Query.Repositories.ProcessingMonths;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using OneOf;

namespace Luxus.Connect.Api.Features.ProcessingMonths;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class ProcessingMonthsController(
    IProcessingMonthQueryRepository queryRepository,
    ICommandSender commandSender,
    AppState appState)
    : ConnectApiController
{
    private const string GET_BY_ID_ROUTE = $"{nameof(ProcessingMonthsController)}.{nameof(GetById)}";

    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListProcessingMonthResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        CancellationToken cancellationToken = default)
    {
        IPagedList<ListProcessingMonthResponse> response = await queryRepository.QueryAsync(
            appState.Organization!.Id,
            pageSearch.ToPageSearch(),
            cancellationToken);

        return Paged(response);
    }

    [HttpGet("{id}", Name = GET_BY_ID_ROUTE)]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetProcessingMonthResponse>> GetById(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        GetProcessingMonthResponse? entity = await queryRepository.LoadAsync(
            appState.Organization!.Id,
            id,
            cancellationToken);

        return entity is null
            ? NotFound(ApiResponse.Fail(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_FOUND))
            : Ok(entity);
    }

    [HttpPost]
    [ProducesResponseType(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetProcessingMonthResponse>> Post(
        [FromBody] CreateProcessingMonthInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<GetProcessingMonthResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id),
            cancellationToken);

        return result.Match<ActionResult<GetProcessingMonthResponse>>(
            entity => CreatedAtRoute(GET_BY_ID_ROUTE, new { id = entity.Id }, entity),
            error => Error(error));
    }

    [HttpPost("{id}/close")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetProcessingMonthResponse>> Close(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        OneOf<GetProcessingMonthResponse, AppError> result = await commandSender.Send(
            new CloseProcessingMonthCommand(appState.Organization!.Id, id),
            cancellationToken);

        return result.Match<ActionResult<GetProcessingMonthResponse>>(
            entity => AcceptedAtRoute(GET_BY_ID_ROUTE, new { id = entity.Id }, entity),
            error => Error(error));
    }

    [HttpPost("{id}/close-contingency")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetProcessingMonthResponse>> CloseInContingency(
        [FromRoute] string id,
        [FromBody] CloseProcessingMonthInContingencyInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<GetProcessingMonthResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id, id),
            cancellationToken);

        return result.Match<ActionResult<GetProcessingMonthResponse>>(
            entity => AcceptedAtRoute(GET_BY_ID_ROUTE, new { id = entity.Id }, entity),
            error => Error(error));
    }
}
