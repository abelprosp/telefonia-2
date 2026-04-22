namespace Luxus.Connect.Domain.PhoneLines.Aggregates;

public enum PhoneLineStatus
{
    INACTIVE = 0,
    ACTIVE = 1,
    CANCELLED = 2,
    SUSPENDED = 3,
    /// <summary>Linha cadastrada mas não presente em fatura importada no período, sem vínculo a outra fatura no mesmo mês de processamento.</summary>
    IN_STOCK = 4,
    /// <summary>Linha vinculada a cliente que deixou de constar no 110D — §2.2 v2 (Aguardando Fatura); não consolida cobrança até voltar a aparecer.</summary>
    AWAITING_INVOICE = 5,
    /// <summary>Venda antecipada / trâmites na operadora — §7 v2 (Em Transição — Aguardando [Tipo]); concilia na primeira aparição no 110D.</summary>
    IN_TRANSITION = 6,
}