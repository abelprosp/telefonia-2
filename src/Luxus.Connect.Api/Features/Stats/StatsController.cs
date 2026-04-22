using Asp.Versioning;
using Luxus.Connect.Contracts.Stats.Responses;
using Luxus.Connect.Infra.Data.Query.Repositories.Stats;
using Luxus.Connect.Infra.Http.Controllers;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace Luxus.Connect.Api.Features.Stats;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class StatsController(IStatsQueryRepository statsQueryRepository)
    : ConnectApiController
{
    [HttpGet("dashboard")]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<DashboardStatsResponse>> Get(CancellationToken cancellationToken = default)
        => Ok(await statsQueryRepository.GetDashboardStats(cancellationToken));
}
