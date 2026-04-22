namespace Luxus.Connect.Domain.Providers.Services;

/// <summary>§10 — proporcionalidade: ((Valor / dias do ciclo) × dias ativos). Referência normativa: ciclo de 30 dias para o denominador.</summary>
public static class ProportionalBillingCalculator
{
    /// <summary>Dias usados como denominador na especificação v2 (independente do calendário civil).</summary>
    public const int DefaultCycleDaysForProRata = 30;

    /// <summary>
    /// Calcula o valor proporcional: <paramref name="monthlyAmount"/> ÷ <paramref name="cycleDays"/> × <paramref name="activeDaysInCycle"/>.
    /// </summary>
    public static decimal CalculateAmount(decimal monthlyAmount, int activeDaysInCycle, int cycleDays = DefaultCycleDaysForProRata)
    {
        if (cycleDays <= 0)
            throw new ArgumentOutOfRangeException(nameof(cycleDays));

        if (activeDaysInCycle < 0)
            throw new ArgumentOutOfRangeException(nameof(activeDaysInCycle));

        int capped = Math.Min(activeDaysInCycle, cycleDays);
        return Math.Round(monthlyAmount / cycleDays * capped, 2, MidpointRounding.AwayFromZero);
    }

    /// <summary>
    /// Conta dias inclusivos na interseção de [cycleStart, cycleEnd] com [serviceStart, serviceEnd].
    /// <paramref name="serviceStart"/> nulo = início do ciclo; <paramref name="serviceEnd"/> nulo = fim do ciclo.
    /// </summary>
    public static int CountInclusiveActiveDaysInCycle(
        DateOnly cycleStart,
        DateOnly cycleEnd,
        DateOnly? serviceStart,
        DateOnly? serviceEnd)
    {
        if (cycleEnd < cycleStart)
            throw new ArgumentException("cycleEnd deve ser >= cycleStart.");

        DateOnly effectiveStart = serviceStart.HasValue ? Max(cycleStart, serviceStart.Value) : cycleStart;
        DateOnly effectiveEnd = serviceEnd.HasValue ? Min(cycleEnd, serviceEnd.Value) : cycleEnd;

        if (effectiveEnd < effectiveStart)
            return 0;

        return effectiveEnd.DayNumber - effectiveStart.DayNumber + 1;
    }

    private static DateOnly Max(DateOnly a, DateOnly b) => a >= b ? a : b;

    private static DateOnly Min(DateOnly a, DateOnly b) => a <= b ? a : b;
}
