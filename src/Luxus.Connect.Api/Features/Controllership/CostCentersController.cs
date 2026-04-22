using Asp.Versioning;
using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
using Luxus.Connect.Contracts.Controllership.Responses;
using Luxus.Connect.Infra.Crosscutting;
using Luxus.Connect.Infra.Data.Query.Repositories.Controllership;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Luxus.Connect.Api.Features.Controllership;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class CostCentersController(
    ICostCenterQueryRepository costCenterQueryRepository,
    AppState appState)
    : ConnectApiController
{
    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListCostCenterResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        CancellationToken cancellationToken = default)
    {
        IPagedList<ListCostCenterResponse> response = await costCenterQueryRepository.QueryAsync(
            appState.Organization!.Id,
            pageSearch.ToPageSearch(),
            cancellationToken);

        return Paged(response);
    }
}
