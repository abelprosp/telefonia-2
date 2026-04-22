using Asp.Versioning;
using Goal.Application.Commands;
using Goal.Infra.Crosscutting.Collections;
using Goal.Infra.Http.Controllers.Requests;
using Goal.Infra.Http.Controllers.Results;
using Goal.Infra.Http.Extensions;
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

namespace Luxus.Connect.Api.Features.Providers;

[ApiController]
[ApiVersion("1")]
[Authorize("admin")]
[Route("v{version:apiVersion}/[controller]")]
public class ProviderInvoicesController(
    IProviderInvoiceQueryRepository providerInvoiceQueryRepository,
    ICommandSender commandSender,
    AppState appState)
    : ConnectApiController
{
    private const string GET_BY_ID_ROUTE = $"{nameof(ProviderInvoicesController)}.{nameof(GetById)}";

    [HttpGet]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<PagedResponse<ListProviderInvoiceResponse>>> Get(
        [FromQuery] PageSearchRequest pageSearch,
        [FromQuery(Name = "processing_month_id")] string? processingMonthId = null,
        CancellationToken cancellationToken = default)
    {
        IPagedList<ListProviderInvoiceResponse> response = await providerInvoiceQueryRepository.QueryAsync(
            appState.Organization!.Id,
            pageSearch.ToPageSearch(),
            processingMonthId,
            cancellationToken);

        return Paged(response);
    }

    [HttpGet("{id}", Name = GET_BY_ID_ROUTE)]
    [ProducesResponseType(StatusCodes.Status200OK)]
    [ProducesResponseType(StatusCodes.Status404NotFound, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<GetProviderInvoiceResponse>> GetById(
        [FromRoute] string id,
        CancellationToken cancellationToken = default)
    {
        GetProviderInvoiceResponse? invoice = await providerInvoiceQueryRepository.LoadAsync(
            appState.Organization!.Id,
            id,
            cancellationToken);

        return invoice is null
            ? NotFound(ApiResponse.Fail(Notifications.Invoices.INVOICE_NOT_FOUND))
            : Ok(invoice);
    }

    [HttpPost]
    [ProducesResponseType(StatusCodes.Status201Created)]
    [ProducesResponseType(StatusCodes.Status400BadRequest, Type = typeof(ApiResponse))]
    [ProducesResponseType(StatusCodes.Status500InternalServerError, Type = typeof(ApiResponse))]
    public async Task<ActionResult<RequestProviderInvoiceImportResponse>> Post(
        [FromBody] ProviderInvoiceImportRequestInput input,
        CancellationToken cancellationToken = default)
    {
        OneOf<RequestProviderInvoiceImportResponse, AppError> result = await commandSender.Send(
            input.ToCommand(appState.Organization!.Id),
            cancellationToken);

        return result.Match(
            model => CreatedAtRoute(GET_BY_ID_ROUTE, new { id = model.Id }, model),
            error => Error(error));
    }
}
