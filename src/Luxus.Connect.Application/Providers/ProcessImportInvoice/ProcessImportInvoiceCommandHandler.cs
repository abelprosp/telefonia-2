using System.Text;
using ConduitR.Abstractions;
using Goal.Application.Commands;
using Goal.Infra.Crosscutting.Localization;
using Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo;
using Luxus.Connect.Contracts.Providers.Commands;
using Luxus.Connect.Contracts.Providers.Events;
using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Luxus.Connect.Domain.Providers.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;
using Luxus.Connect.Infra.Crosscutting.Bus;
using Luxus.Connect.Infra.Crosscutting.Constants;
using Luxus.Connect.Infra.Crosscutting.Errors;
using Luxus.Connect.Infra.Crosscutting.Extensions;
using Luxus.Connect.Infra.Crosscutting.ObjectStorage;
using Luxus.Connect.Infra.Data;
using Microsoft.Extensions.Logging;
using OneOf;
using OneOf.Types;
using static Luxus.Connect.Application.Invoices.ImportInvoice.Parsers.Vivo.VivoTextInvoiceParser;

namespace Luxus.Connect.Application.Providers.ProcessImportInvoice;

internal sealed class ProcessImportInvoiceCommandHandler(
    ILogger<ProcessImportInvoiceCommandHandler> logger,
    IAppUnitOfWork uow,
    IObjectStorageClient objectStorage,
    IBusPublisher busPublisher)
    : ICommandHandler<ImportInvoiceCommand, OneOf<None, AppError>>, IRequestHandler<ImportInvoiceCommand, OneOf<None, AppError>>
{
    private static readonly Encoding Latin1Encoding = Encoding.GetEncoding("ISO-8859-1");
    private sealed class ImportMatrixCounters
    {
        public int NewLinesInStockCount { get; set; }
        public int TransitionedToActiveCount { get; set; }
        public int AbsentToAwaitingInvoiceCount { get; set; }
        public int AbsentToInactiveStockCount { get; set; }
        public int StructuralWarningsCount { get; set; }
    }

    private sealed record ResolvedImportContext(
        ContractingCompany ContractingCompany,
        ProviderAccount ProviderAccount,
        ProcessingMonth ProcessingMonth,
        BillingCycle BillingCycle);

    private sealed record ResolvedImportCustomer(
        Customer? Customer,
        string? SourceDocument,
        IReadOnlyCollection<string> SuggestedCustomerIds);

    public async ValueTask<OneOf<None, AppError>> Handle(ImportInvoiceCommand command, CancellationToken cancellationToken)
    {
        ProviderInvoiceImportRequest? entity =
            await uow.InvoiceImportRequests.GetAsync(command.ImportRequestId, cancellationToken);

        if (entity is null)
        {
            return new BusinessRuleError(Notifications.InvoiceImports.IMPORT_REQUEST_NOT_FOUND);
        }

        try
        {
            if (entity.Status != ProviderInvoiceImportRequestStatus.PENDING)
            {
                return new BusinessRuleError(Notifications.InvoiceImports.IMPORT_REQUEST_NOT_PENDING);
            }

            entity.MarkProcessing();

            await uow.CommitAsync(cancellationToken);

            await using ObjectStorageObject objectPayload = await objectStorage
                .GetObjectAsync(entity.StorageBucket, entity.StorageObjectKey, cancellationToken);

            OneOf<None, AppError> processResult = await ProcessAsync(
                entity,
                objectPayload.Stream,
                cancellationToken);

            if (processResult.IsError())
            {
                AppError error = processResult.GetError();

                entity.MarkFailed(string.Join("; ", error.Notifications.Select(n => n.Message)));

                await uow.CommitAsync(cancellationToken);
                return error;
            }

            entity.MarkCompleted();

            await uow.CommitAsync(cancellationToken);
        }
        catch (Exception ex)
        {
            logger.LogError(ex, "Invoice import {ImportRequestId} failed. Details: {Details}", entity.Id, ex);

            string message = ex switch
            {
                ObjectStorageObjectNotFoundException osx =>
                    $"Object not found in storage: {osx.BucketName}/{osx.ObjectKey}",
                _ => ex.Message,
            };

            entity.MarkFailed(message);

            await uow.CommitAsync(cancellationToken);
        }

        return default(None);
    }

    private async Task<OneOf<None, AppError>> ProcessAsync(
        ProviderInvoiceImportRequest importRequest,
        Stream content,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(importRequest);
        ArgumentNullException.ThrowIfNull(content);

        Provider? provider = await uow.Providers.GetAsync(importRequest.ProviderId, cancellationToken);

        if (provider is null)
        {
            return new BusinessRuleError(Notifications.Providers.PROVIDER_NOT_FOUND);
        }

        List<object> contentParsed = await ParseFile(content, cancellationToken);

        Line010DHeader header = GetParsedHeader(contentParsed);

        OneOf<ResolvedImportContext, AppError> resolved =
            await ResolveImportContextAsync(provider, importRequest, contentParsed, header, cancellationToken);

        if (resolved.IsError())
        {
            return resolved.GetError();
        }

        ResolvedImportContext ctx = resolved.GetSuccess();
        var counters = new ImportMatrixCounters();
        OneOf<ResolvedImportCustomer, AppError> resolvedCustomer =
            await ResolveCustomerFrom011DAsync(provider, ctx.ContractingCompany, contentParsed, cancellationToken);
        if (resolvedCustomer.IsError())
        {
            return resolvedCustomer.GetError();
        }

        ResolvedImportCustomer importCustomer = resolvedCustomer.GetSuccess();

        ProviderInvoiceDuplication duplicate = await uow.Invoices.FindDuplicateByBusinessKeyAsync(
            ctx.ProviderAccount.AccountNumber,
            ctx.ContractingCompany.Id,
            ctx.ProcessingMonth.Id,
            header.DueDate,
            cancellationToken);

        if (duplicate == ProviderInvoiceDuplication.Duplicate)
            return new BusinessRuleError(Notifications.Invoices.INVOICE_DUPLICATE_SAME_PROCESSING_MONTH);

        HashSet<string> numbersInFile = BuildNormalizedPhoneNumbersFrom110D(contentParsed);

        var planCache = new Dictionary<string, ProviderPlan>(StringComparer.Ordinal);
        var planServiceCache = new Dictionary<string, ProviderPlanService>(StringComparer.Ordinal);

        ProviderInvoice invoice = await GetInvoice(
            ctx.ProviderAccount,
            ctx.ContractingCompany,
            ctx.ProcessingMonth,
            ctx.BillingCycle,
            provider,
            header,
            numbersInFile,
            contentParsed,
            planCache,
            planServiceCache,
            importCustomer,
            counters,
            cancellationToken);

        OneOf<None, AppError> structuralValidation = ValidateImportedLinesStructuralConsistency(invoice);
        if (structuralValidation.IsError())
        {
            return structuralValidation.GetError();
        }

        await PublishMatrixAlertAsync(importRequest, ctx, counters, cancellationToken);

        logger.LogInformation(
            "Invoice import matrix summary for request {ImportRequestId}. Provider={ProviderId}, Account={ProviderAccountId}, ProcessingMonth={ProcessingMonthId}, NewInStock={NewLinesInStockCount}, TransitionToActive={TransitionedToActiveCount}, AbsentToAwaiting={AbsentToAwaitingInvoiceCount}, AbsentToInactiveStock={AbsentToInactiveStockCount}, StructuralWarnings={StructuralWarningsCount}",
            importRequest.Id,
            ctx.ContractingCompany.ProviderId,
            ctx.ProviderAccount.Id,
            ctx.ProcessingMonth.Id,
            counters.NewLinesInStockCount,
            counters.TransitionedToActiveCount,
            counters.AbsentToAwaitingInvoiceCount,
            counters.AbsentToInactiveStockCount,
            counters.StructuralWarningsCount);

        uow.Invoices.Add(invoice);

        return default(None);
    }

    private static OneOf<None, AppError> ValidateImportedLinesStructuralConsistency(ProviderInvoice invoice)
    {
        foreach (PhoneLine line in invoice.PhoneLines)
        {
            // Regra estrutural (§3.1, escopo atual): linha presente em fatura não pode terminar
            // em estados incompatíveis com destino operacional.
            if (line.Status is PhoneLineStatus.INACTIVE
                or PhoneLineStatus.CANCELLED
                or PhoneLineStatus.SUSPENDED)
            {
                return new BusinessRuleError(Notifications.Invoices.INVOICE_IMPORTED_LINE_ORPHAN_DESTINATION);
            }

            if (line.Status == PhoneLineStatus.IN_TRANSITION && line.TransitionSubStatus is null)
            {
                return new BusinessRuleError(Notifications.Invoices.INVOICE_IMPORTED_LINE_ORPHAN_DESTINATION);
            }
        }

        // Mismatch cliente-da-linha vs cliente-da-fatura depende do vínculo explícito PhoneLine<->Customer
        // (épico 3). Por isso, nesta fase validamos apenas coerência de estado operacional.
        return default(None);
    }

    private async Task PublishMatrixAlertAsync(
        ProviderInvoiceImportRequest importRequest,
        ResolvedImportContext ctx,
        ImportMatrixCounters counters,
        CancellationToken cancellationToken)
    {
        var evt = new InvoiceImportMatrixAlertEvent(
            importRequest.Id,
            ctx.ContractingCompany.ProviderId,
            ctx.ProviderAccount.Id,
            ctx.ProcessingMonth.Id,
            counters.NewLinesInStockCount,
            counters.TransitionedToActiveCount,
            counters.AbsentToAwaitingInvoiceCount,
            counters.AbsentToInactiveStockCount,
            counters.StructuralWarningsCount,
            "Invoice import matrix transitions processed.");

        await busPublisher.Publish(evt, cancellationToken);
    }

    private static HashSet<string> BuildNormalizedPhoneNumbersFrom110D(List<object> contentParsed)
    {
        var set = new HashSet<string>(StringComparer.Ordinal);

        foreach (Line110DAccountLineDetail item in contentParsed.OfType<Line110DAccountLineDetail>())
        {
            string phoneNumber = item.PhoneNumber.NormalizeDigitsOnly();

            if (phoneNumber.Length > 0)
                set.Add(phoneNumber);
        }

        return set;
    }

    private async Task<ProviderPlan?> ResolvePlanForImportAsync(
        Provider op,
        string planCode,
        string nameForNewPlan,
        Dictionary<string, ProviderPlan> planCache,
        CancellationToken cancellationToken)
    {
        if (string.IsNullOrWhiteSpace(planCode))
        {
            return null;
        }

        string cacheKey = op.Id + "\u001f" + planCode;
        if (planCache.TryGetValue(cacheKey, out ProviderPlan? cached))
        {
            return cached;
        }

        ProviderPlan? plan = await uow.ProviderPlans.GetByProviderAndCode(op.Id, planCode, cancellationToken);

        if (plan is null)
        {
            plan = ProviderPlan.Create(op, nameForNewPlan, planCode);
            await uow.ProviderPlans.AddAsync(plan, cancellationToken);
        }

        planCache[cacheKey] = plan;
        return plan;
    }

    private async Task<ProviderPlanService> ResolvePlanServiceForImportAsync(
        ProviderPlan plan,
        string serviceName,
        Dictionary<string, ProviderPlanService> planServiceCache,
        CancellationToken cancellationToken)
    {
        string cacheKey = plan.Id + "\u001f" + serviceName;
        if (planServiceCache.TryGetValue(cacheKey, out ProviderPlanService? cached))
        {
            return cached;
        }

        ProviderPlanService? planService =
            await uow.PlanServices.GetByProviderAndNameAsync(plan.Id, serviceName, cancellationToken);

        if (planService is null)
        {
            planService = ProviderPlanService.Create(plan, serviceName, false);
            await uow.PlanServices.AddAsync(planService, cancellationToken);
        }

        planServiceCache[cacheKey] = planService;
        return planService;
    }

    private async Task<ProviderInvoice> GetInvoice(
        ProviderAccount providerAccount,
        ContractingCompany contractingCompany,
        ProcessingMonth processingMonth,
        BillingCycle billingCycle,
        Provider provider,
        Line010DHeader header,
        HashSet<string> numbersInFile,
        List<object> contentParsed,
        Dictionary<string, ProviderPlan> planCache,
        Dictionary<string, ProviderPlanService> planServiceCache,
        ResolvedImportCustomer importCustomer,
        ImportMatrixCounters counters,
        CancellationToken cancellationToken)
    {
        var invoice = ProviderInvoice.Create(
            providerAccount,
            contractingCompany,
            processingMonth,
            billingCycle,
            header.IssueDate,
            header.DueDate,
            header.TotalAmount);

        invoice.SetSubtotals(
            header.SubtotalServices,
            header.SubtotalUsageExceeded,
            0,
            0,
            0);

        await ExtractPlansAndServices(
            provider,
            contentParsed,
            invoice,
            planCache,
            planServiceCache,
            cancellationToken);

        await ExtractInvoiceUsage(provider, contentParsed, invoice, planCache, cancellationToken);
        ExtractInvoiceExcess(contentParsed, invoice);
        await ExtractInvoiceLines(provider, providerAccount, invoice, contentParsed, planCache, importCustomer, counters, cancellationToken);
        await ApplyAbsentLinesPostProcessing(providerAccount.Id, invoice, numbersInFile, counters, cancellationToken);

        return invoice;
    }

    private async Task ApplyAbsentLinesPostProcessing(
        string providerAccountId,
        ProviderInvoice invoice,
        HashSet<string> numbersInFile,
        ImportMatrixCounters counters,
        CancellationToken cancellationToken)
    {
        IEnumerable<PhoneLine> accountLines =
            await uow.PhoneLines.ListByAccountIdAsync(providerAccountId, cancellationToken);

        foreach (PhoneLine line in accountLines)
        {
            string lineNumberKey = line.Number.NormalizeDigitsOnly();

            if (string.IsNullOrWhiteSpace(lineNumberKey) || numbersInFile.Contains(lineNumberKey))
            {
                continue;
            }

            bool hasActiveCustomer = line.ActiveCustomerLink is not null;

            if (!hasActiveCustomer)
            {
                line.RecordLastInvoice(invoice);
                line.MarkInactiveInStockWhenAbsentFromInvoice();
                counters.AbsentToInactiveStockCount++;
                continue;
            }

            line.RecordLastInvoice(invoice);
            line.MarkAsAwaitingInvoice();
            counters.AbsentToAwaitingInvoiceCount++;
        }
    }

    private async Task ExtractInvoiceLines(
        Provider provider,
        ProviderAccount providerAccount,
        ProviderInvoice invoice,
        List<object> contentParsed,
        Dictionary<string, ProviderPlan> planCache,
        ResolvedImportCustomer importCustomer,
        ImportMatrixCounters counters,
        CancellationToken cancellationToken)
    {
        List<Line110DAccountLineDetail> linesParsed = contentParsed
            .OfType<Line110DAccountLineDetail>()
            .ToList()!;

        foreach (Line110DAccountLineDetail item in linesParsed
            .GroupBy(i => i.PhoneNumber.NormalizeDigitsOnly())
            .Select(g => g.First()))
        {
            string numberKey = item.PhoneNumber.NormalizeDigitsOnly();

            if (string.IsNullOrEmpty(numberKey))
                continue;

            ProviderPlan? plan = await ResolvePlanForImportAsync(
                provider,
                item.PlanName,
                item.PlanName,
                planCache,
                cancellationToken);

            if (plan is null)
            {
                continue;
            }

            PhoneLine? phoneLine = await uow.PhoneLines.GetByNumberAsync(numberKey, cancellationToken);
            bool isNewLine = false;

            if (phoneLine is not null && phoneLine.ProviderAccountId != providerAccount.Id)
            {
                throw new InvalidOperationException(
                    $"Linha '{numberKey}' já vinculada a outra conta ({phoneLine.ProviderAccountId}).");
            }

            if (phoneLine is null)
            {
                phoneLine = PhoneLine.Create(plan, providerAccount, numberKey);
                await uow.PhoneLines.AddAsync(phoneLine, cancellationToken);
                isNewLine = true;
            }

            if (phoneLine.ActiveCustomerLink is null && importCustomer.Customer is not null)
            {
                phoneLine.AssignCustomer(importCustomer.Customer, invoice.IssueDate);
                importCustomer.Customer.AddProviderLink(provider, invoice.IssueDate);
                importCustomer.Customer.Reactivate();
            }

            // §3.2/§3.3 (estoque): snapshot financeiro da linha no mês corrente.
            // Com os dados disponíveis no 110D, usamos o total da linha como base inicial.
            phoneLine.SetCostSnapshot(item.LineTotal, item.LineTotal);

            bool hasActiveCustomer = phoneLine.ActiveCustomerLink is not null;
            bool transitionedToActive = hasActiveCustomer && phoneLine.Status == PhoneLineStatus.IN_TRANSITION;
            invoice.LinkPlanLine(phoneLine);
            if (hasActiveCustomer)
            {
                phoneLine.ApplyImportedLinePresence(invoice, invoice.IssueDate);
                if (transitionedToActive && phoneLine.Status == PhoneLineStatus.ACTIVE)
                {
                    counters.TransitionedToActiveCount++;
                }
            }
            else
            {
                phoneLine.RecordLastInvoice(invoice);
                phoneLine.MarkAsInStock();

                if (isNewLine)
                {
                    counters.NewLinesInStockCount++;
                }
            }
        }

        if (importCustomer.Customer is null && importCustomer.SuggestedCustomerIds.Count > 0)
        {
            counters.StructuralWarningsCount++;
            logger.LogWarning(
                "Invoice import {InvoiceId} has ambiguous customer match from 011D document {TaxId}. Suggestions={SuggestedCustomerIds}",
                invoice.Id,
                importCustomer.SourceDocument,
                importCustomer.SuggestedCustomerIds);
        }
    }

    private async Task<OneOf<ResolvedImportCustomer, AppError>> ResolveCustomerFrom011DAsync(
        Provider provider,
        ContractingCompany contractingCompany,
        List<object> contentParsed,
        CancellationToken cancellationToken)
    {
        Line011DCustomer? parsedCustomer = contentParsed
            .OfType<Line011DCustomer>()
            .FirstOrDefault();

        if (parsedCustomer is null)
        {
            return new ResolvedImportCustomer(null, null, []);
        }

        string taxId = parsedCustomer.Document.NormalizeDigitsOnly();
        if (string.IsNullOrWhiteSpace(taxId))
        {
            return new ResolvedImportCustomer(null, null, []);
        }

        if (taxId.Length is not (11 or 14))
        {
            return new BusinessRuleError(Notifications.InvoiceImports.CUSTOMER_DOCUMENT_INVALID_FOR_IMPORT);
        }

        var matches = (await uow.Customers.ListByDocumentAsync(provider.OrganizationId, taxId, cancellationToken))
            .Where(c => c.Active && c.HasActiveProvider(provider.Id))
            .ToList();

        if (matches.Count == 1)
        {
            Customer matchedCustomer = matches[0];
            string? customerCnpj = matchedCustomer.Documents
                .Where(d => d.DocumentType == CustomerDocumentType.CNPJ)
                .Select(d => d.Number)
                .FirstOrDefault();

            if (!string.IsNullOrWhiteSpace(customerCnpj) && customerCnpj != contractingCompany.TaxId)
            {
                return new BusinessRuleError(Notifications.Customers.CUSTOMER_CONTRACTING_COMPANY_MISMATCH);
            }

            return new ResolvedImportCustomer(matchedCustomer, taxId, []);
        }

        if (matches.Count > 1)
        {
            return new ResolvedImportCustomer(null, taxId, [.. matches.Select(c => c.Id)]);
        }

        return new ResolvedImportCustomer(null, taxId, []);
    }

    private void ExtractInvoiceExcess(List<object> contentParsed, ProviderInvoice invoice)
    {
        List<string> types = ["052W", "052E", "052D"];

        List<LineRecord> extraUsage = contentParsed
            .OfType<LineRecord>()
            .Where(p => types.Contains(p.RecordType))
            .ToList()!;

        Line052WExtraUsageHeader? lastHeader = null;
        Line052EExtraLocation? lastLocation = null;
        Line052DExtraUsageDetail? lastDetail = null;
        ProviderInvoiceItem? headerParent = null;
        ProviderInvoiceItem? locationParent = null;

        foreach (LineRecord item in extraUsage)
        {
            if (item.RecordType == "052W")
            {
                lastHeader = (Line052WExtraUsageHeader)item;

                headerParent = invoice.AddItem(
                    lastHeader.Description,
                    0,
                    0,
                    ProviderInvoiceItemType.EXTRA_HEADER);

                continue;
            }

            if (item.RecordType == "052E")
            {
                lastLocation = (Line052EExtraLocation)item;

                if (string.IsNullOrWhiteSpace(lastLocation.Location))
                {
                    locationParent = headerParent;
                    continue;
                }

                locationParent = invoice.AddItem(
                    lastLocation.Location,
                    0,
                    0,
                    ProviderInvoiceItemType.EXTRA_LOCATION,
                    headerParent);

                continue;
            }

            if (item.RecordType == "052D")
            {
                lastDetail = (Line052DExtraUsageDetail)item;

                invoice.AddItem(
                    lastDetail.Description,
                    lastDetail.Quantity,
                    lastDetail.Amount,
                    ProviderInvoiceItemType.EXTRA_DETAIL,
                    locationParent);
            }
        }
    }

    private async Task ExtractInvoiceUsage(
        Provider provider,
        List<object> contentParsed,
        ProviderInvoice invoice,
        Dictionary<string, ProviderPlan> planCache,
        CancellationToken cancellationToken)
    {
        var franchisesParsed = contentParsed
            .OfType<Line050GService>()
            .OrderBy(g => g.ServiceName)
            .GroupBy(g => new { g.ServiceName, g.Unity })
            .ToList()!;

        foreach (var franchiseItem in franchisesParsed)
        {
            Line050GService sample = franchiseItem.First();
            ProviderPlan? plan = await ResolvePlanForImportAsync(
                provider,
                sample.PlanCode,
                sample.ServiceName,
                planCache,
                cancellationToken);

            invoice.AddItem(
                franchiseItem.Key.ServiceName,
                franchiseItem.Sum(i => i.Quantity),
                franchiseItem.Sum(i => i.Total),
                ProviderInvoiceItemType.USAGE,
                null,
                franchiseItem.Sum(i => i.Franchise),
                franchiseItem.Sum(i => i.Used),
                Enum.Parse<InvoiceItemUnit>(franchiseItem.Key.Unity));

            invoice.AddService(
                plan,
                franchiseItem.Key.ServiceName,
                franchiseItem.Sum(i => i.Quantity),
                franchiseItem.Sum(i => i.Total),
                franchiseItem.Sum(i => i.Franchise),
                franchiseItem.Sum(i => i.Used),
                Enum.Parse<InvoiceItemUnit>(franchiseItem.Key.Unity));
        }
    }

    private async Task ExtractPlansAndServices(
        Provider provider,
        List<object> contentParsed,
        ProviderInvoice invoice,
        Dictionary<string, ProviderPlan> planCache,
        Dictionary<string, ProviderPlanService> planServiceCache,
        CancellationToken cancellationToken)
    {
        List<string> planSummaries = contentParsed
            .OfType<Line050IPlanSummary>()
            .Where(s => s.Flags != "NNDUMMY")
            .Select(p => p.PlanCode)
            .ToList()!;

        List<Line050HService> plansParsed = contentParsed
            .OfType<Line050HService>()
            .Where(s => planSummaries.Contains(s.PlanCode) && (s.Flags== "M" || string.IsNullOrWhiteSpace(s.Flags)))
            .ToList()!;

        ProviderPlan? plan;
        ProviderPlanService? planService;
        ProviderInvoiceItem? parent;

        foreach (Line050HService item in plansParsed)
        {
            plan = await ResolvePlanForImportAsync(
                provider,
                item.PlanCode,
                item.ServiceName,
                planCache,
                cancellationToken);

            if (plan is null)
            {
                continue;
            }

            parent = invoice.AddItem(
                plan.Name,
                item.Quantity,
                item.Total,
                ProviderInvoiceItemType.PLAN);

            invoice.AddService(
                plan,
                plan.Name,
                item.Quantity,
                item.Total,
                null,
                null,
                null);

            List<Line050HService> servicesParsed = contentParsed
                .OfType<Line050HService>()
                .Where(s => s.PlanCode ==  item.PlanCode && s.Flags == "A")
                .ToList()!;

            foreach (Line050HService? serviceItem in servicesParsed)
            {
                if (string.IsNullOrWhiteSpace(serviceItem.ServiceName))
                {
                    continue;
                }

                planService = await ResolvePlanServiceForImportAsync(
                    plan,
                    serviceItem.ServiceName,
                    planServiceCache,
                    cancellationToken);

                invoice.AddItem(
                    serviceItem.ServiceName,
                    serviceItem.Quantity,
                    serviceItem.Total,
                    ProviderInvoiceItemType.SERVICE,
                    parent);

                invoice.AddService(
                    plan,
                    serviceItem.ServiceName,
                    serviceItem.Quantity,
                    serviceItem.Total,
                    null,
                    null,
                    null);
            }
        }
    }

    private static async Task<List<object>> ParseFile(Stream content, CancellationToken cancellationToken)
    {
        await using MemoryStream buffer = await ReadAllLimitedAsync(content, cancellationToken);

        if (buffer.Length == 0)
            throw new InvalidOperationException("Imported file is empty.");

        byte[] raw = buffer.ToArray();
        string text = DecodeTextBestEffort(raw);

        var contentParsed = VivoTextInvoiceParser
            .Parse(text)
            .ToList();

        return contentParsed;
    }

    private async Task<OneOf<ResolvedImportContext, AppError>> ResolveImportContextAsync(
        Provider provider,
        ProviderInvoiceImportRequest importRequest,
        List<object> contentParsed,
        Line010DHeader header,
        CancellationToken cancellationToken)
    {
        Line011DCustomer? parsedCustomer = contentParsed
            .OfType<Line011DCustomer>()
            .FirstOrDefault();

        if (parsedCustomer is null)
        {
            throw new InvalidOperationException("Arquivo sem registro de cliente (011D).");
        }

        string taxId = parsedCustomer.Document.NormalizeDigitsOnly();
        if (taxId.Length is not (11 or 14))
        {
            return new BusinessRuleError(Notifications.InvoiceImports.CUSTOMER_DOCUMENT_INVALID_FOR_IMPORT);
        }

        ContractingCompany? company;
        if (taxId.Length == 14)
        {
            company = await uow.ContractingCompanies.GetByProviderAndTaxIdAsync(
                provider.Id,
                taxId,
                cancellationToken);

            if (company is null)
            {
                company = ContractingCompany.Create(
                    provider,
                    ResolveLegalNameForImportContractingCompany(parsedCustomer),
                    taxId);

                await uow.ContractingCompanies.AddAsync(company, cancellationToken);
            }
        }
        else
        {
            var cpfCustomers = (await uow.Customers.ListByDocumentAsync(
                provider.OrganizationId,
                taxId,
                cancellationToken))
                .Where(c => c.HasActiveProvider(provider.Id))
                .ToList();

            if (cpfCustomers.Count != 1)
            {
                return new BusinessRuleError(Notifications.InvoiceImports.CPF_REQUIRES_EXISTING_CUSTOMER_FOR_IMPORT);
            }

            Customer customer = cpfCustomers[0];

            string? customerCnpj = customer.Documents
                .Where(d => d.DocumentType == CustomerDocumentType.CNPJ)
                .Select(d => d.Number)
                .FirstOrDefault();

            if (string.IsNullOrWhiteSpace(customerCnpj))
            {
                return new BusinessRuleError(Notifications.Customers.CUSTOMER_CONTRACTING_COMPANY_MISMATCH);
            }

            company = await uow.ContractingCompanies.GetByProviderAndTaxIdAsync(
                provider.Id,
                customerCnpj,
                cancellationToken);

            if (company is null)
            {
                return new BusinessRuleError(Notifications.InvoiceImports.CONTRACTING_COMPANY_NOT_FOUND_FOR_FILE);
            }
        }

        ProviderAccount? account = await uow.ProviderAccounts.GetByContractingCompanyAndAccountNumber(
            company.Id,
            header.AccountNumber,
            cancellationToken);

        if (account is null)
        {
            account = ProviderAccount.Create(
                company,
                header.AccountNumber);

            await uow.ProviderAccounts.AddAsync(account, cancellationToken);
        }

        BillingCycle? cycle = await uow.BillingCycles.GetByCodeAsync(
            provider.Id,
            header.ReferenceMonth,
            cancellationToken);

        if (cycle is null)
        {
            if (await uow.ProcessingMonths.ExistsClosedIntersectingDateRangeAsync(
                    provider.OrganizationId,
                    provider.Id,
                    header.BillingStartDate,
                    header.BillingEndDate,
                    cancellationToken))
            {
                return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_RETROACTIVE_CHANGE_BLOCKED);
            }

            cycle = BillingCycle.Create(
                provider,
                header.ReferenceMonth,
                header.BillingEndDate.ToString("MMMM yyyy", ApplicationCultures.Portugues),
                header.BillingStartDate,
                header.BillingEndDate);

            await uow.BillingCycles.AddAsync(cycle, cancellationToken);
        }

        ProcessingMonth? processingMonth = await uow.ProcessingMonths.GetByIdAsync(
            provider.OrganizationId,
            importRequest.ProcessingMonthId,
            cancellationToken);

        if (processingMonth is null)
        {
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_FOUND);
        }

        if (processingMonth.ProviderId != provider.Id)
        {
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_PROVIDER_MISMATCH);
        }

        if (processingMonth.Status != ProcessingMonthStatus.OPEN)
        {
            return new BusinessRuleError(Notifications.ProcessingMonths.PROCESSING_MONTH_NOT_OPEN);
        }

        return new ResolvedImportContext(company, account, processingMonth, cycle);
    }

    private static string ResolveLegalNameForImportContractingCompany(Line011DCustomer parsedCustomer)
    {
        string legal = parsedCustomer.LegalName?.Trim() ?? string.Empty;
        string name = parsedCustomer.Name?.Trim() ?? string.Empty;

        if (!string.IsNullOrWhiteSpace(legal))
        {
            return legal;
        }

        if (!string.IsNullOrWhiteSpace(name))
        {
            return name;
        }

        return "Empresa não identificada";
    }

    private static Line010DHeader GetParsedHeader(List<object> contentParsed)
    {
        Line010DHeader? header = contentParsed
            .OfType<Line010DHeader>()
            .FirstOrDefault();

        if (header is null)
        {
            throw new InvalidOperationException("Arquivo sem registro de cabeçalho de fatura (010D).");
        }

        if (string.IsNullOrWhiteSpace(header.AccountNumber))
        {
            throw new InvalidOperationException("Número da conta (accountNumber) está vazio.");
        }

        return header;
    }

    private static string DecodeTextBestEffort(byte[] raw)
    {
        try
        {
            return Latin1Encoding.GetString(raw);
        }
        catch (DecoderFallbackException)
        {
            return Encoding.UTF8.GetString(raw);
        }
    }

    private static async Task<MemoryStream> ReadAllLimitedAsync(
        Stream source,
        CancellationToken cancellationToken)
    {
        byte[] buffer = new byte[81920];
        var aggregate = new MemoryStream();

        int read;
        while ((read = await source.ReadAsync(buffer.AsMemory(0, buffer.Length), cancellationToken)) > 0)
        {
            await aggregate.WriteAsync(buffer.AsMemory(0, read), cancellationToken);
        }

        aggregate.Position = 0;

        return aggregate;
    }
}
