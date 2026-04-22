namespace Luxus.Connect.Domain.ProcessingMonths;

/// <summary>Interseção entre intervalo de datas e um mês civil (ano/mês).</summary>
public static class ProcessingMonthDateRange
{
    public static bool IntersectsCivilMonth(DateOnly rangeStart, DateOnly rangeEnd, int year, int month)
    {
        if (rangeEnd < rangeStart)
            return false;

        DateOnly monthStart = new(year, month, 1);
        DateOnly monthEnd = monthStart.AddMonths(1).AddDays(-1);

        return !(rangeEnd < monthStart || rangeStart > monthEnd);
    }
}
