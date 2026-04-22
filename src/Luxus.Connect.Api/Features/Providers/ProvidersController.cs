using Asp.Versioning;
using Goal.Application.Commands;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Contracts.Providers.Inputs;
using Luxus.Connect.Contracts.Providers.Responses;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Data.Query.Repositories.Providers;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using OneOf;
using OneOf.Types;

namespace Luxus.Connect.Api.Features.Providers;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class ProvidersController(
    IProviderQueryRepository providerQueryRepository,
    ICommandSender commandSender,
    AppState appState)
    : ConnectApiController
{
    private const string GET_BY_ID_ROUTE = $"{nameof(ProvidersController)}.{nameof(GetById)}";

    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListProvidersResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        CancellationToken cancellationToken = default)
    {
        return Paged(
            await providerQueryRepository.QueryAsync(
                appState.Organization!.Id,
                pageSearch.ToPageSearch(), cancellationToken));
    }

    [HttpGet("{id}", Name = GET_BY_ID_ROUTE)]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetProviderResponse>> GetById([FromRoute] string id, CancellationToken cancellationToken = default)
    {
        GetProviderResponse? op = await providerQueryRepository.GetWithDetailsAsync(appState.Organization!.Id, id, cancellationToken);

        return op is null
            ? NotFound(ApiResponse.Fail(Notifications.Providers.PROVIDER_NOT_FOUND))
            : Ok(op);
    }

    [HttpPost]
    [ProducesResponseType(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<CreateProviderResponse>> Post(
        [FromBody] CreateProviderInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<CreateProviderResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id),
            cancellationToken);

        return result
            .Match<ActionResult<CreateProviderResponse>>(
                op => CreatedAtRoute(GET_BY_ID_ROUTE, new { id = op.Id }, op),
                error => Error(error)
            );
    }

    [HttpPatch("{id}")]
    [ProducesResponseType(StatusCodes.Status202Accepted)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status409Conflict, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult> Patch([FromRoute] string id, [FromBody] UpdateProviderInput input)
    {
        OneOf<None, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id, id));

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
        OneOf<None, AppError> result = await commandSender.Send<OneOf<None, AppError>>(
            new InactivateProviderCommand(appState.Organization!.Id, id));

        return result
            .Match(
                _ => Accepted(),
                Error
            );
    }
}
