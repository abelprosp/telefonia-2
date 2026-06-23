using Luxus.Connect.Domain.BillingCycles.Aggregates;
using Luxus.Connect.Domain.Customers.Aggregates;
using Luxus.Connect.Domain.PhoneLines.Aggregates;
using Luxus.Connect.Domain.Providers.Enums;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations;

/// <inheritdoc />
public partial class _01 : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.CreateTable(
            name: "AuditLogs",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ChangeType = table.Column<string>(type: "character varying(16)", maxLength: 16, nullable: false),
                EntityName = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                KeyValues = table.Column<string>(type: "text", nullable: false),
                ChangedBy = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: true),
                OldValues = table.Column<string>(type: "text", nullable: true),
                NewValues = table.Column<string>(type: "text", nullable: true),
                Timestamp = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_AuditLogs", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "CostCenters",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Name = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: false),
                Description = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_CostCenters", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "Providers",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Name = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                Slug = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                Active = table.Column<bool>(type: "boolean", nullable: false, defaultValue: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_Providers", x => x.Id);
            });

        migrationBuilder.CreateTable(
            name: "BillingCycles",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Code = table.Column<string>(type: "character varying(20)", maxLength: 20, nullable: false),
                Name = table.Column<string>(type: "character varying(100)", maxLength: 100, nullable: false),
                StartDate = table.Column<DateOnly>(type: "date", nullable: false),
                EndDate = table.Column<DateOnly>(type: "date", nullable: false),
                Status = table.Column<BillingCycleStatus>(type: "billing_cycle_status", nullable: false),
                ClosedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: true),
                ClosedBy = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_BillingCycles", x => x.Id);
                table.ForeignKey(
                    name: "FK_BillingCycles_Providers_ProviderId",
                    column: x => x.ProviderId,
                    principalTable: "Providers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ContractingCompanies",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                LegalName = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: false),
                TaxId = table.Column<string>(type: "character varying(14)", maxLength: 14, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ContractingCompanies", x => x.Id);
                table.ForeignKey(
                    name: "FK_ContractingCompanies_Providers_ProviderId",
                    column: x => x.ProviderId,
                    principalTable: "Providers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Restrict);
            });

        migrationBuilder.CreateTable(
            name: "Customers",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Type = table.Column<CustomerType>(type: "customer_type", nullable: false),
                Name = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                LegalName = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: true),
                BirthOrOpeningDate = table.Column<DateOnly>(type: "date", nullable: true),
                Active = table.Column<bool>(type: "boolean", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_Customers", x => x.Id);
                table.ForeignKey(
                    name: "FK_Customers_Providers_ProviderId",
                    column: x => x.ProviderId,
                    principalTable: "Providers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderInvoiceImportRequests",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                StorageBucket = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                StorageObjectKey = table.Column<string>(type: "character varying(2048)", maxLength: 2048, nullable: false),
                OriginalFileName = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: true),
                Status = table.Column<int>(type: "integer", nullable: false),
                Error = table.Column<string>(type: "character varying(8000)", maxLength: 8000, nullable: true),
                CompletedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: true),
                CreatedBy = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderInvoiceImportRequests", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceImportRequests_Providers_ProviderId",
                    column: x => x.ProviderId,
                    principalTable: "Providers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Restrict);
            });

        migrationBuilder.CreateTable(
            name: "ProviderPlans",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Name = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                Code = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderPlans", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderPlans_Providers_ProviderId",
                    column: x => x.ProviderId,
                    principalTable: "Providers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderAccounts",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ContractingCompanyId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                AccountNumber = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderAccounts", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderAccounts_ContractingCompanies_ContractingCompanyId",
                    column: x => x.ContractingCompanyId,
                    principalTable: "ContractingCompanies",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Restrict);
            });

        migrationBuilder.CreateTable(
            name: "CustomerAddresses",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                CustomerId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Street = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                Number = table.Column<string>(type: "character varying(16)", maxLength: 16, nullable: false),
                Neighborhood = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: false),
                Complement = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: true),
                City = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: false),
                State = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                ZipCode = table.Column<string>(type: "character varying(10)", maxLength: 10, nullable: false),
                Country = table.Column<string>(type: "character varying(32)", maxLength: 32, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_CustomerAddresses", x => x.Id);
                table.ForeignKey(
                    name: "FK_CustomerAddresses_Customers_CustomerId",
                    column: x => x.CustomerId,
                    principalTable: "Customers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "CustomerDocuments",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                CustomerId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                DocumentType = table.Column<CustomerDocumentType>(type: "customer_document_type", nullable: false),
                Number = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_CustomerDocuments", x => x.Id);
                table.ForeignKey(
                    name: "FK_CustomerDocuments_Customers_CustomerId",
                    column: x => x.CustomerId,
                    principalTable: "Customers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderPlanServices",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderPlanId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Name = table.Column<string>(type: "character varying(256)", maxLength: 256, nullable: false),
                Active = table.Column<bool>(type: "boolean", nullable: false, defaultValue: true),
                Recurring = table.Column<bool>(type: "boolean", nullable: false),
                Price = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderPlanServices", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderPlanServices_ProviderPlans_ProviderPlanId",
                    column: x => x.ProviderPlanId,
                    principalTable: "ProviderPlans",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "BillingCycleProviderAccounts",
            columns: table => new
            {
                BillingCyclesId = table.Column<string>(type: "character varying(36)", nullable: false),
                ProviderAccountsId = table.Column<string>(type: "character varying(36)", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_BillingCycleProviderAccounts", x => new { x.BillingCyclesId, x.ProviderAccountsId });
                table.ForeignKey(
                    name: "FK_BillingCycleProviderAccounts_BillingCycles_BillingCyclesId",
                    column: x => x.BillingCyclesId,
                    principalTable: "BillingCycles",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_BillingCycleProviderAccounts_ProviderAccounts_ProviderAccou~",
                    column: x => x.ProviderAccountsId,
                    principalTable: "ProviderAccounts",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderInvoices",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderAccountId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ContractingCompanyId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                BillingCycleId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                CostCenterId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ParentInvoiceId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ReferenceYear = table.Column<int>(type: "integer", nullable: false),
                ReferenceMonth = table.Column<int>(type: "integer", nullable: false),
                IssueDate = table.Column<DateOnly>(type: "date", nullable: false),
                DueDate = table.Column<DateOnly>(type: "date", nullable: false),
                TotalAmount = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                Status = table.Column<ProviderInvoiceStatus>(type: "provider_invoice_status", nullable: false),
                SubtotalServices = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                SubtotalUsage = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                SubtotalTaxes = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                SubtotalDiscounts = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                SubtotalInstallments = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderInvoices", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderInvoices_BillingCycles_BillingCycleId",
                    column: x => x.BillingCycleId,
                    principalTable: "BillingCycles",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Restrict);
                table.ForeignKey(
                    name: "FK_ProviderInvoices_ContractingCompanies_ContractingCompanyId",
                    column: x => x.ContractingCompanyId,
                    principalTable: "ContractingCompanies",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_ProviderInvoices_CostCenters_CostCenterId",
                    column: x => x.CostCenterId,
                    principalTable: "CostCenters",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.SetNull);
                table.ForeignKey(
                    name: "FK_ProviderInvoices_ProviderAccounts_ProviderAccountId",
                    column: x => x.ProviderAccountId,
                    principalTable: "ProviderAccounts",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_ProviderInvoices_ProviderInvoices_ParentInvoiceId",
                    column: x => x.ParentInvoiceId,
                    principalTable: "ProviderInvoices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "PhoneLines",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderPlanId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderAccountId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                CostCenterId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: true),
                LastInvoiceId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: true),
                TitularLineId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: true),
                Number = table.Column<string>(type: "character varying(20)", maxLength: 20, nullable: false),
                LineClassification = table.Column<LineClassification>(type: "line_classification", nullable: false),
                Status = table.Column<PhoneLineStatus>(type: "phone_line_status", nullable: false),
                TransitionSubStatus = table.Column<TransitionSubStatus>(type: "transition_sub_status", nullable: true),
                TransitionStartedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: true),
                ActivationDate = table.Column<DateOnly>(type: "date", nullable: true),
                CancellationDate = table.Column<DateOnly>(type: "date", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_PhoneLines", x => x.Id);
                table.ForeignKey(
                    name: "FK_PhoneLines_CostCenters_CostCenterId",
                    column: x => x.CostCenterId,
                    principalTable: "CostCenters",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.SetNull);
                table.ForeignKey(
                    name: "FK_PhoneLines_PhoneLines_TitularLineId",
                    column: x => x.TitularLineId,
                    principalTable: "PhoneLines",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_PhoneLines_ProviderAccounts_ProviderAccountId",
                    column: x => x.ProviderAccountId,
                    principalTable: "ProviderAccounts",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_PhoneLines_ProviderInvoices_LastInvoiceId",
                    column: x => x.LastInvoiceId,
                    principalTable: "ProviderInvoices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_PhoneLines_ProviderPlans_ProviderPlanId",
                    column: x => x.ProviderPlanId,
                    principalTable: "ProviderPlans",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.SetNull);
            });

        migrationBuilder.CreateTable(
            name: "ProviderInvoiceItems",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                InvoiceId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ParentId = table.Column<string>(type: "character varying(36)", nullable: true),
                Description = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: false),
                Quantity = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: false),
                TotalPrice = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                ItemType = table.Column<ProviderInvoiceItemType>(type: "provider_invoice_item_type", nullable: false),
                QuotaAmount = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: true),
                ConsumedAmount = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: true),
                Unit = table.Column<InvoiceItemUnit>(type: "invoice_item_unit", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderInvoiceItems", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceItems_ProviderInvoiceItems_ParentId",
                    column: x => x.ParentId,
                    principalTable: "ProviderInvoiceItems",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceItems_ProviderInvoices_InvoiceId",
                    column: x => x.InvoiceId,
                    principalTable: "ProviderInvoices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderInvoiceServices",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                InvoiceId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                PlanId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Description = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: false),
                Quantity = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: false),
                TotalPrice = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: false),
                QuotaAmount = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: true),
                ConsumedAmount = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: true),
                Unit = table.Column<InvoiceItemUnit>(type: "invoice_item_unit", nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderInvoiceServices", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceServices_ProviderInvoices_InvoiceId",
                    column: x => x.InvoiceId,
                    principalTable: "ProviderInvoices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceServices_ProviderPlans_PlanId",
                    column: x => x.PlanId,
                    principalTable: "ProviderPlans",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "PhoneLineServices",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                PhoneLineId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderPlanServiceId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Name = table.Column<string>(type: "text", nullable: false),
                Code = table.Column<string>(type: "text", nullable: false),
                Recurring = table.Column<bool>(type: "boolean", nullable: false),
                Price = table.Column<decimal>(type: "numeric(18,2)", precision: 18, scale: 2, nullable: true),
                Active = table.Column<bool>(type: "boolean", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_PhoneLineServices", x => x.Id);
                table.ForeignKey(
                    name: "FK_PhoneLineServices_PhoneLines_PhoneLineId",
                    column: x => x.PhoneLineId,
                    principalTable: "PhoneLines",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_PhoneLineServices_ProviderPlanServices_ProviderPlanServiceId",
                    column: x => x.ProviderPlanServiceId,
                    principalTable: "ProviderPlanServices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderInvoicePhoneLines",
            columns: table => new
            {
                PhoneLinesId = table.Column<string>(type: "character varying(36)", nullable: false),
                ProviderInvoicesId = table.Column<string>(type: "character varying(36)", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderInvoicePhoneLines", x => new { x.PhoneLinesId, x.ProviderInvoicesId });
                table.ForeignKey(
                    name: "FK_ProviderInvoicePhoneLines_PhoneLines_PhoneLinesId",
                    column: x => x.PhoneLinesId,
                    principalTable: "PhoneLines",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_ProviderInvoicePhoneLines_ProviderInvoices_ProviderInvoices~",
                    column: x => x.ProviderInvoicesId,
                    principalTable: "ProviderInvoices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateTable(
            name: "ProviderInvoiceQuotaSharing",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                InvoiceId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                PhoneLineId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Description = table.Column<string>(type: "character varying(512)", maxLength: 512, nullable: false),
                ConsumedAmount = table.Column<decimal>(type: "numeric(18,4)", precision: 18, scale: 4, nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProviderInvoiceQuotaSharing", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceQuotaSharing_PhoneLines_PhoneLineId",
                    column: x => x.PhoneLineId,
                    principalTable: "PhoneLines",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_ProviderInvoiceQuotaSharing_ProviderInvoices_InvoiceId",
                    column: x => x.InvoiceId,
                    principalTable: "ProviderInvoices",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateIndex(
            name: "IX_AuditLogs_EntityName",
            table: "AuditLogs",
            column: "EntityName");

        migrationBuilder.CreateIndex(
            name: "IX_AuditLogs_EntityName_ChangedBy",
            table: "AuditLogs",
            columns: new[] { "EntityName", "ChangedBy" });

        migrationBuilder.CreateIndex(
            name: "IX_BillingCycleProviderAccounts_ProviderAccountsId",
            table: "BillingCycleProviderAccounts",
            column: "ProviderAccountsId");

        migrationBuilder.CreateIndex(
            name: "IX_BillingCycles_OrganizationId",
            table: "BillingCycles",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_BillingCycles_OrganizationId_Code",
            table: "BillingCycles",
            columns: new[] { "OrganizationId", "Code" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_BillingCycles_OrganizationId_Name",
            table: "BillingCycles",
            columns: new[] { "OrganizationId", "Name" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_BillingCycles_ProviderId",
            table: "BillingCycles",
            column: "ProviderId");

        migrationBuilder.CreateIndex(
            name: "IX_ContractingCompanies_ProviderId_TaxId",
            table: "ContractingCompanies",
            columns: new[] { "ProviderId", "TaxId" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_CostCenters_OrganizationId",
            table: "CostCenters",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_CostCenters_OrganizationId_Name",
            table: "CostCenters",
            columns: new[] { "OrganizationId", "Name" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_CustomerAddresses_CustomerId",
            table: "CustomerAddresses",
            column: "CustomerId");

        migrationBuilder.CreateIndex(
            name: "IX_CustomerDocuments_CustomerId",
            table: "CustomerDocuments",
            column: "CustomerId");

        migrationBuilder.CreateIndex(
            name: "IX_CustomerDocuments_DocumentType",
            table: "CustomerDocuments",
            column: "DocumentType");

        migrationBuilder.CreateIndex(
            name: "IX_Customers_OrganizationId",
            table: "Customers",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_Customers_OrganizationId_ProviderId_LegalName",
            table: "Customers",
            columns: new[] { "OrganizationId", "ProviderId", "LegalName" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_Customers_OrganizationId_ProviderId_Name",
            table: "Customers",
            columns: new[] { "OrganizationId", "ProviderId", "Name" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_Customers_ProviderId",
            table: "Customers",
            column: "ProviderId");

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLines_CostCenterId",
            table: "PhoneLines",
            column: "CostCenterId");

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLines_LastInvoiceId",
            table: "PhoneLines",
            column: "LastInvoiceId");

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLines_ProviderAccountId",
            table: "PhoneLines",
            column: "ProviderAccountId");

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLines_ProviderPlanId",
            table: "PhoneLines",
            column: "ProviderPlanId");

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLines_Status_TransitionStartedAt",
            table: "PhoneLines",
            columns: new[] { "Status", "TransitionStartedAt" });

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLines_TitularLineId",
            table: "PhoneLines",
            column: "TitularLineId");

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLineServices_PhoneLineId_ProviderPlanServiceId",
            table: "PhoneLineServices",
            columns: new[] { "PhoneLineId", "ProviderPlanServiceId" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_PhoneLineServices_ProviderPlanServiceId",
            table: "PhoneLineServices",
            column: "ProviderPlanServiceId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderAccounts_AccountNumber",
            table: "ProviderAccounts",
            column: "AccountNumber");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderAccounts_ContractingCompanyId_AccountNumber",
            table: "ProviderAccounts",
            columns: new[] { "ContractingCompanyId", "AccountNumber" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceImportRequests_OrganizationId",
            table: "ProviderInvoiceImportRequests",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceImportRequests_OrganizationId_ProviderId",
            table: "ProviderInvoiceImportRequests",
            columns: new[] { "OrganizationId", "ProviderId" });

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceImportRequests_ProviderId",
            table: "ProviderInvoiceImportRequests",
            column: "ProviderId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceItems_InvoiceId",
            table: "ProviderInvoiceItems",
            column: "InvoiceId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceItems_ParentId",
            table: "ProviderInvoiceItems",
            column: "ParentId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoicePhoneLines_ProviderInvoicesId",
            table: "ProviderInvoicePhoneLines",
            column: "ProviderInvoicesId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceQuotaSharing_InvoiceId",
            table: "ProviderInvoiceQuotaSharing",
            column: "InvoiceId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceQuotaSharing_PhoneLineId",
            table: "ProviderInvoiceQuotaSharing",
            column: "PhoneLineId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_BillingCycleId",
            table: "ProviderInvoices",
            column: "BillingCycleId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_ContractingCompanyId",
            table: "ProviderInvoices",
            column: "ContractingCompanyId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_CostCenterId",
            table: "ProviderInvoices",
            column: "CostCenterId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_ParentInvoiceId",
            table: "ProviderInvoices",
            column: "ParentInvoiceId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_ProviderAccountId_ContractingCompanyId_Bil~",
            table: "ProviderInvoices",
            columns: new[] { "ProviderAccountId", "ContractingCompanyId", "BillingCycleId", "DueDate" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceServices_InvoiceId",
            table: "ProviderInvoiceServices",
            column: "InvoiceId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceServices_PlanId",
            table: "ProviderInvoiceServices",
            column: "PlanId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderPlans_ProviderId_Code",
            table: "ProviderPlans",
            columns: new[] { "ProviderId", "Code" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_ProviderPlanServices_ProviderPlanId",
            table: "ProviderPlanServices",
            column: "ProviderPlanId");

        migrationBuilder.CreateIndex(
            name: "IX_Providers_OrganizationId",
            table: "Providers",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_Providers_OrganizationId_Active",
            table: "Providers",
            columns: new[] { "OrganizationId", "Active" });

        migrationBuilder.CreateIndex(
            name: "IX_Providers_OrganizationId_Slug",
            table: "Providers",
            columns: new[] { "OrganizationId", "Slug" },
            unique: true);
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropTable(
            name: "AuditLogs");

        migrationBuilder.DropTable(
            name: "BillingCycleProviderAccounts");

        migrationBuilder.DropTable(
            name: "CustomerAddresses");

        migrationBuilder.DropTable(
            name: "CustomerDocuments");

        migrationBuilder.DropTable(
            name: "PhoneLineServices");

        migrationBuilder.DropTable(
            name: "ProviderInvoiceImportRequests");

        migrationBuilder.DropTable(
            name: "ProviderInvoiceItems");

        migrationBuilder.DropTable(
            name: "ProviderInvoicePhoneLines");

        migrationBuilder.DropTable(
            name: "ProviderInvoiceQuotaSharing");

        migrationBuilder.DropTable(
            name: "ProviderInvoiceServices");

        migrationBuilder.DropTable(
            name: "Customers");

        migrationBuilder.DropTable(
            name: "ProviderPlanServices");

        migrationBuilder.DropTable(
            name: "PhoneLines");

        migrationBuilder.DropTable(
            name: "ProviderInvoices");

        migrationBuilder.DropTable(
            name: "ProviderPlans");

        migrationBuilder.DropTable(
            name: "BillingCycles");

        migrationBuilder.DropTable(
            name: "CostCenters");

        migrationBuilder.DropTable(
            name: "ProviderAccounts");

        migrationBuilder.DropTable(
            name: "ContractingCompanies");

        migrationBuilder.DropTable(
            name: "Providers");
    }
}
