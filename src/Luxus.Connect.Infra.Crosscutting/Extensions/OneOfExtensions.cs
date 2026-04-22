using Luxus.Connect.Infra.Crosscutting.Errors;
using OneOf;

namespace Luxus.Connect.Infra.Crosscutting.Extensions;

public static class OneOfExtensions
{
    public static bool IsSuccess<TResult, TError>(this OneOf<TResult, TError> obj)
        where TError : AppError
        => obj.IsT0;

    public static bool IsError<TResult, TError>(this OneOf<TResult, TError> obj)
        where TError : AppError
        => obj.IsT1;

    public static TResult GetSuccess<TResult, TError>(this OneOf<TResult, TError> obj)
        where TError : AppError
        => obj.AsT0;

    public static TError GetError<TResult, TError>(this OneOf<TResult, TError> obj)
        where TError : AppError
        => obj.AsT1;
}