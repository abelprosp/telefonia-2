using System.Linq.Expressions;
using System.Reflection;

namespace Luxus.Connect.Infra.Crosscutting.TextParsing;

/// <summary>Extrai o literal comparado à propriedade <c>RecordType</c> em expressões do tipo <c>r => r.RecordType == "010D"</c>.</summary>
public static class RecordTypeExpression
{
    private const string RecordTypePropertyName = "RecordType";

    /// <summary>
    /// Interpreta <paramref name="predicate"/> como igualdade entre a propriedade <c>RecordType</c> e uma constante string.
    /// </summary>
    public static bool TryGetExpectedRecordType<TRecord>(Expression<Func<TRecord, bool>> predicate, out string recordType)
    {
        recordType = null!;
        Expression body = predicate.Body;

        if (body is UnaryExpression unary && unary.NodeType == ExpressionType.Convert)
            body = unary.Operand;

        if (body is not BinaryExpression binary || binary.NodeType != ExpressionType.Equal)
            return false;

        return TryEqual(binary.Left, binary.Right, out recordType)
            || TryEqual(binary.Right, binary.Left, out recordType);
    }

    private static bool TryEqual(Expression a, Expression b, out string recordType)
    {
        recordType = null!;
        if (!TryGetRecordTypeMember(a, out _))
            (a, b) = (b, a);

        if (!TryGetRecordTypeMember(a, out _))
            return false;

        if (b is ConstantExpression c && c.Value is string s)
        {
            recordType = s;
            return true;
        }

        if (b is UnaryExpression conv && conv.NodeType == ExpressionType.Convert && conv.Operand is ConstantExpression c2 && c2.Value is string s2)
        {
            recordType = s2;
            return true;
        }

        return false;
    }

    private static bool TryGetRecordTypeMember(Expression expression, out MemberInfo? member)
    {
        member = null;
        Expression e = expression;
        if (e is UnaryExpression u && u.NodeType == ExpressionType.Convert)
            e = u.Operand;

        if (e is MemberExpression me && me.Member is PropertyInfo pi && pi.Name == RecordTypePropertyName)
        {
            member = pi;
            return true;
        }

        return false;
    }
}
