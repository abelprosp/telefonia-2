using Asp.Versioning;
using Goal.Application.Commands;
using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
using Luxus.Connect.Contracts.BillingCycles.Inputs;
using Luxus.Connect.Contracts.BillingCycles.Responses;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data.Query.Repositories.BillingCycles;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Api.Features.BillingCycles;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class BillingCyclesController(
    IBillingCycleQueryRepository billingCycleQueryRepository,
    ICommandSender commandSender,
    AppState appState)
    : ConnectApiController
{
    private const string GET_BY_ID_ROUTE = $"{nameof(BillingCyclesController)}.{nameof(GetById)}";

    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListBillingCycleResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        CancellationToken cancellationToken = default)
    {
        IPagedList<ListBillingCycleResponse> response = await billingCycleQueryRepository.QueryAsync(
            appState.Organization!.Id,
            pageSearch.ToPageSearch(),
            cancellationToken);

        return Paged(response);
    }

    [HttpGet("{id}", Name = GET_BY_ID_ROUTE)]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetBillingCycleResponse>> GetById(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        GetBillingCycleResponse? cycle = await billingCycleQueryRepository.LoadAsync(
            appState.Organization!.Id,
            id,
            cancellationToken);

        return cycle is null
            ? NotFound(ApiResponse.Fail(Notifications.BillingCycles.BILLING_CYCLE_NOT_FOUND))
            : Ok(cycle);
    }

    [HttpPost]
    [ProducesResponseType(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<CreateBillingCycleResponse>> Post(
        [FromBody] CreateBillingCycleInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<CreateBillingCycleResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id),
            cancellationToken);

        return result
            .Match<ActionResult<CreateBillingCycleResponse>>(
                cycle => CreatedAtRoute(GET_BY_ID_ROUTE, new { id = cycle.Id }, cycle),
                error => Error(error)
            );
    }

    [HttpPatch("{id}")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult> Patch([FromRoute] string id, [FromBody] UpdateBillingCycleInput input)
    {
        OneOf<None, AppError> result = await commandSender.Send(input.ToCommand(appState.Organization!.Id, id));

        return result
            .Match(
                _ => AcceptedAtRoute(GET_BY_ID_ROUTE, new { id }),
                Error
            );
    }
}
