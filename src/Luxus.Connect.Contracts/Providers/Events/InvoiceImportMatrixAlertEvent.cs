using Goal.Application.Events;

namespace Luxus.Connect.Contracts.Providers.Events;

public sealed class InvoiceImportMatrixAlertEvent(
    string aggregateId,
    string providerId,
    string providerAccountId,
    string processingMonthId,
    int newLinesInStockCount,
    int transitionedToActiveCount,
    int absentToAwaitingInvoiceCount,
    int absentToInactiveStockCount,
    int structuralWarningsCount,
    string summaryMessage)
    : Event(aggregateId, nameof(InvoiceImportMatrixAlertEvent))
{
    private InvoiceImportMatrixAlertEvent()
        : this(string.Empty, string.Empty, string.Empty, string.Empty, 0, 0, 0, 0, 0, string.Empty)
    {
    }

    public const string RabbitMqRoutingKey = "providers.invoice_import.matrix_alert";

    public string ProviderId { get; } = providerId;
    public string ProviderAccountId { get; } = providerAccountId;
    public string ProcessingMonthId { get; } = processingMonthId;
    public int NewLinesInStockCount { get; } = newLinesInStockCount;
    public int TransitionedToActiveCount { get; } = transitionedToActiveCount;
    public int AbsentToAwaitingInvoiceCount { get; } = absentToAwaitingInvoiceCount;
    public int AbsentToInactiveStockCount { get; } = absentToInactiveStockCount;
    public int StructuralWarningsCount { get; } = structuralWarningsCount;
    public string SummaryMessage { get; } = summaryMessage;
}
