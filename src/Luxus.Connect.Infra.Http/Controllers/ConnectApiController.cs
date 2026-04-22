using Goal.Infra.Http.Controllers;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Microsoft.AspNetCore.Mvc;

namespace Luxus.Connect.Infra.Http.Controllers;

public class ConnectApiController : ApiController
{
    protected ActionResult Error(AppError error)
    {
        var responseMap = new Dictionary<Type, Func<ActionResult>>
        {
            { typeof(BusinessRuleError), () => Conflict(ApiResponse.Fail(error.Notifications)) },
            { typeof(InputValidationError), () => BadRequest(ApiResponse.Fail(error.Notifications)) },
            { typeof(ResourceNotFoundError), () => NotFound(ApiResponse.Fail(error.Notifications)) }
        };

        return responseMap.TryGetValue(error.GetType(), out Func<ActionResult>? response)
            ? response.Invoke()
            : InternalServerError(ApiResponse.Fail(error.Notifications));
    }
}
