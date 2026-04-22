namespace Luxus.Connect.Domain.Providers.Enums;

/// <summary>
/// Resultado da verificação de duplicidade por chave de negócio:
/// conta (nuvem) + empresa contratante + mês de processamento + vencimento.
/// </summary>
public enum ProviderInvoiceDuplication
{
    None,
    Duplicate
}
