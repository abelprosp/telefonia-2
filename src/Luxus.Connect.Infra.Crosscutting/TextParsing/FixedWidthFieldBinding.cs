using System.Linq.Expressions;
using System.Reflection;

namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

internal sealed class FixedWidthFieldBinding<TRecord> where TRecord : class, new()
{
    public required PropertyInfo Property { get; init; }
    public required int Offset { get; init; }
    public required int Length { get; init; }
    public string? Format { get; init; }
    public required Action<TRecord, string> Assign { get; init; }

    public static FixedWidthFieldBinding<TRecord> Create(
        Expression<Func<TRecord, object?>> property,
        int offset,
        int length,
        string? format,
        bool trimFieldValues)
    {
        PropertyInfo pi = GetProperty(property);
        Action<TRecord, string> assign = FieldValueBinder.CreateSetter<TRecord>(pi, format, trimFieldValues);
        return new FixedWidthFieldBinding<TRecord>
        {
            Property = pi,
            Offset = offset,
            Length = length,
            Format = format,
            Assign = assign,
        };
    }

    /// <summary>Sobrecarga para propriedades de tipo valor (DateOnly, int, etc.) sem boxing em <c>object</c>.</summary>
    public static FixedWidthFieldBinding<TRecord> Create<TProperty>(
        Expression<Func<TRecord, TProperty>> property,
        int offset,
        int length,
        string? format,
        bool trimFieldValues)
    {
        PropertyInfo pi = GetPropertyFromTyped(property);
        Action<TRecord, string> assign = FieldValueBinder.CreateSetter<TRecord>(pi, format, trimFieldValues);
        return new FixedWidthFieldBinding<TRecord>
        {
            Property = pi,
            Offset = offset,
            Length = length,
            Format = format,
            Assign = assign,
        };
    }

    private static PropertyInfo GetProperty(Expression<Func<TRecord, object?>> property)
    {
        Expression body = property.Body;
        if (body is UnaryExpression unary && unary.NodeType == ExpressionType.Convert)
            body = unary.Operand;

        if (body is not MemberExpression member || member.Member is not PropertyInfo pi)
            throw new ArgumentException("A expressão deve apontar para uma propriedade.", nameof(property));

        return pi;
    }

    private static PropertyInfo GetPropertyFromTyped<TProperty>(Expression<Func<TRecord, TProperty>> property)
    {
        Expression body = property.Body;
        if (body is UnaryExpression unary && unary.NodeType == ExpressionType.Convert)
            body = unary.Operand;

        if (body is not MemberExpression member || member.Member is not PropertyInfo pi)
            throw new ArgumentException("A expressão deve apontar para uma propriedade.", nameof(property));

        return pi;
    }

    public void Apply(string line, TRecord target)
    {
        if (Offset < 0 || Length <= 0)
            return;

        if (Offset >= line.Length)
        {
            Assign(target, string.Empty);
            return;
        }

        int len = Math.Min(Length, line.Length - Offset);
        string slice = line.Substring(Offset, len);
        Assign(target, slice);
    }
}
