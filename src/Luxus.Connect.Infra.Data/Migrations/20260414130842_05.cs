using Luxus.Connect.Domain.ProcessingMonths.Aggregates;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations;

/// <inheritdoc />
public partial class _05 : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropIndex(
            name: "IX_ProviderInvoices_ProviderAccountId_ContractingCompanyId_Bil~",
            table: "ProviderInvoices");

        migrationBuilder.DropColumn(
            name: "ReferenceMonth",
            table: "ProviderInvoices");

        migrationBuilder.DropColumn(
            name: "ReferenceYear",
            table: "ProviderInvoices");

        migrationBuilder.AlterDatabase()
            .Annotation("Npgsql:Enum:billing_cycle_status", "closed,open")
            .Annotation("Npgsql:Enum:customer_document_type", "cnh,cnpj,cpf,municipal_registration,other,rg,state_registration")
            .Annotation("Npgsql:Enum:customer_type", "pf,pj")
            .Annotation("Npgsql:Enum:exceedance_charge_type", "mirroed")
            .Annotation("Npgsql:Enum:invoice_item_unit", "gb,kb,mb,min,sms,tb")
            .Annotation("Npgsql:Enum:line_classification", "dependent,normal,other,titular")
            .Annotation("Npgsql:Enum:phone_line_status", "active,awaiting_invoice,cancelled,in_stock,in_transition,inactive,suspended")
            .Annotation("Npgsql:Enum:processing_month_status", "closed,open")
            .Annotation("Npgsql:Enum:provider_invoice_item_type", "discount,extra_detail,extra_header,extra_location,other,plan,service,usage")
            .Annotation("Npgsql:Enum:provider_invoice_status", "cancelled,draft,overdue,paid,pending")
            .Annotation("Npgsql:Enum:service_application_type", "addon,plan,service")
            .Annotation("Npgsql:Enum:service_availability_rule", "always,custom,cycle_only")
            .Annotation("Npgsql:Enum:service_type", "data,other,roaming,sms,subscription")
            .Annotation("Npgsql:Enum:transition_sub_status", "none,pending_activation,pending_cancellation,pending_portability,pending_pp,pending_tt")
            .OldAnnotation("Npgsql:Enum:billing_cycle_status", "closed,open")
            .OldAnnotation("Npgsql:Enum:customer_document_type", "cnh,cnpj,cpf,municipal_registration,other,rg,state_registration")
            .OldAnnotation("Npgsql:Enum:customer_type", "pf,pj")
            .OldAnnotation("Npgsql:Enum:exceedance_charge_type", "mirroed")
            .OldAnnotation("Npgsql:Enum:invoice_item_unit", "gb,kb,mb,min,sms,tb")
            .OldAnnotation("Npgsql:Enum:line_classification", "dependent,normal,other,titular")
            .OldAnnotation("Npgsql:Enum:phone_line_status", "active,awaiting_invoice,cancelled,in_stock,in_transition,inactive,suspended")
            .OldAnnotation("Npgsql:Enum:provider_invoice_item_type", "discount,extra_detail,extra_header,extra_location,other,plan,service,usage")
            .OldAnnotation("Npgsql:Enum:provider_invoice_status", "cancelled,draft,overdue,paid,pending")
            .OldAnnotation("Npgsql:Enum:service_application_type", "addon,plan,service")
            .OldAnnotation("Npgsql:Enum:service_availability_rule", "always,custom,cycle_only")
            .OldAnnotation("Npgsql:Enum:service_type", "data,other,roaming,sms,subscription")
            .OldAnnotation("Npgsql:Enum:transition_sub_status", "none,pending_activation,pending_cancellation,pending_portability,pending_pp,pending_tt");

        migrationBuilder.AddColumn<string>(
            name: "ProcessingMonthId",
            table: "ProviderInvoices",
            type: "character varying(36)",
            maxLength: 36,
            nullable: false,
            defaultValue: "");

        migrationBuilder.AddColumn<string>(
            name: "ProcessingMonthId",
            table: "ProviderInvoiceImportRequests",
            type: "character varying(36)",
            maxLength: 36,
            nullable: false,
            defaultValue: "");

        migrationBuilder.CreateTable(
            name: "ProcessingMonths",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Year = table.Column<int>(type: "integer", nullable: false),
                Month = table.Column<int>(type: "integer", nullable: false),
                DisplayName = table.Column<string>(type: "character varying(128)", maxLength: 128, nullable: false),
                Status = table.Column<ProcessingMonthStatus>(type: "processing_month_status", nullable: false),
                ClosedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: true),
                ClosedBy = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: true),
                ClosedInContingency = table.Column<bool>(type: "boolean", nullable: false),
                ContingencyJustification = table.Column<string>(type: "character varying(4000)", maxLength: 4000, nullable: true)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_ProcessingMonths", x => x.Id);
                table.ForeignKey(
                    name: "FK_ProcessingMonths_Providers_ProviderId",
                    column: x => x.ProviderId,
                    principalTable: "Providers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Restrict);
            });

        migrationBuilder.CreateTable(
            name: "CustomerProcessingMonthManualReleases",
            columns: table => new
            {
                Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                OrganizationId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                CustomerId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                ProcessingMonthId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                Justification = table.Column<string>(type: "character varying(4000)", maxLength: 4000, nullable: false),
                ReleasedByUserId = table.Column<string>(type: "character varying(64)", maxLength: 64, nullable: false),
                ReleasedAt = table.Column<DateTimeOffset>(type: "timestamp with time zone", nullable: false)
            },
            constraints: table =>
            {
                table.PrimaryKey("PK_CustomerProcessingMonthManualReleases", x => x.Id);
                table.ForeignKey(
                    name: "FK_CustomerProcessingMonthManualReleases_Customers_CustomerId",
                    column: x => x.CustomerId,
                    principalTable: "Customers",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
                table.ForeignKey(
                    name: "FK_CustomerProcessingMonthManualReleases_ProcessingMonths_Proc~",
                    column: x => x.ProcessingMonthId,
                    principalTable: "ProcessingMonths",
                    principalColumn: "Id",
                    onDelete: ReferentialAction.Cascade);
            });

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_ProcessingMonthId",
            table: "ProviderInvoices",
            column: "ProcessingMonthId");

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_ProviderAccountId_ContractingCompanyId_Pro~",
            table: "ProviderInvoices",
            columns: new[] { "ProviderAccountId", "ContractingCompanyId", "ProcessingMonthId", "DueDate" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoiceImportRequests_ProcessingMonthId",
            table: "ProviderInvoiceImportRequests",
            column: "ProcessingMonthId");

        migrationBuilder.CreateIndex(
            name: "IX_CustomerProcessingMonthManualReleases_CustomerId_Processing~",
            table: "CustomerProcessingMonthManualReleases",
            columns: new[] { "CustomerId", "ProcessingMonthId" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_CustomerProcessingMonthManualReleases_OrganizationId",
            table: "CustomerProcessingMonthManualReleases",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_CustomerProcessingMonthManualReleases_ProcessingMonthId",
            table: "CustomerProcessingMonthManualReleases",
            column: "ProcessingMonthId");

        migrationBuilder.CreateIndex(
            name: "IX_ProcessingMonths_OrganizationId",
            table: "ProcessingMonths",
            column: "OrganizationId");

        migrationBuilder.CreateIndex(
            name: "IX_ProcessingMonths_OrganizationId_ProviderId_Year_Month",
            table: "ProcessingMonths",
            columns: new[] { "OrganizationId", "ProviderId", "Year", "Month" },
            unique: true);

        migrationBuilder.CreateIndex(
            name: "IX_ProcessingMonths_ProviderId",
            table: "ProcessingMonths",
            column: "ProviderId");

        migrationBuilder.AddForeignKey(
            name: "FK_ProviderInvoiceImportRequests_ProcessingMonths_ProcessingMo~",
            table: "ProviderInvoiceImportRequests",
            column: "ProcessingMonthId",
            principalTable: "ProcessingMonths",
            principalColumn: "Id",
            onDelete: ReferentialAction.Cascade);

        migrationBuilder.AddForeignKey(
            name: "FK_ProviderInvoices_ProcessingMonths_ProcessingMonthId",
            table: "ProviderInvoices",
            column: "ProcessingMonthId",
            principalTable: "ProcessingMonths",
            principalColumn: "Id",
            onDelete: ReferentialAction.Cascade);
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropForeignKey(
            name: "FK_ProviderInvoiceImportRequests_ProcessingMonths_ProcessingMo~",
            table: "ProviderInvoiceImportRequests");

        migrationBuilder.DropForeignKey(
            name: "FK_ProviderInvoices_ProcessingMonths_ProcessingMonthId",
            table: "ProviderInvoices");

        migrationBuilder.DropTable(
            name: "CustomerProcessingMonthManualReleases");

        migrationBuilder.DropTable(
            name: "ProcessingMonths");

        migrationBuilder.DropIndex(
            name: "IX_ProviderInvoices_ProcessingMonthId",
            table: "ProviderInvoices");

        migrationBuilder.DropIndex(
            name: "IX_ProviderInvoices_ProviderAccountId_ContractingCompanyId_Pro~",
            table: "ProviderInvoices");

        migrationBuilder.DropIndex(
            name: "IX_ProviderInvoiceImportRequests_ProcessingMonthId",
            table: "ProviderInvoiceImportRequests");

        migrationBuilder.DropColumn(
            name: "ProcessingMonthId",
            table: "ProviderInvoices");

        migrationBuilder.DropColumn(
            name: "ProcessingMonthId",
            table: "ProviderInvoiceImportRequests");

        migrationBuilder.AlterDatabase()
            .Annotation("Npgsql:Enum:billing_cycle_status", "closed,open")
            .Annotation("Npgsql:Enum:customer_document_type", "cnh,cnpj,cpf,municipal_registration,other,rg,state_registration")
            .Annotation("Npgsql:Enum:customer_type", "pf,pj")
            .Annotation("Npgsql:Enum:exceedance_charge_type", "mirroed")
            .Annotation("Npgsql:Enum:invoice_item_unit", "gb,kb,mb,min,sms,tb")
            .Annotation("Npgsql:Enum:line_classification", "dependent,normal,other,titular")
            .Annotation("Npgsql:Enum:phone_line_status", "active,awaiting_invoice,cancelled,in_stock,in_transition,inactive,suspended")
            .Annotation("Npgsql:Enum:provider_invoice_item_type", "discount,extra_detail,extra_header,extra_location,other,plan,service,usage")
            .Annotation("Npgsql:Enum:provider_invoice_status", "cancelled,draft,overdue,paid,pending")
            .Annotation("Npgsql:Enum:service_application_type", "addon,plan,service")
            .Annotation("Npgsql:Enum:service_availability_rule", "always,custom,cycle_only")
            .Annotation("Npgsql:Enum:service_type", "data,other,roaming,sms,subscription")
            .Annotation("Npgsql:Enum:transition_sub_status", "none,pending_activation,pending_cancellation,pending_portability,pending_pp,pending_tt")
            .OldAnnotation("Npgsql:Enum:billing_cycle_status", "closed,open")
            .OldAnnotation("Npgsql:Enum:customer_document_type", "cnh,cnpj,cpf,municipal_registration,other,rg,state_registration")
            .OldAnnotation("Npgsql:Enum:customer_type", "pf,pj")
            .OldAnnotation("Npgsql:Enum:exceedance_charge_type", "mirroed")
            .OldAnnotation("Npgsql:Enum:invoice_item_unit", "gb,kb,mb,min,sms,tb")
            .OldAnnotation("Npgsql:Enum:line_classification", "dependent,normal,other,titular")
            .OldAnnotation("Npgsql:Enum:phone_line_status", "active,awaiting_invoice,cancelled,in_stock,in_transition,inactive,suspended")
            .OldAnnotation("Npgsql:Enum:processing_month_status", "closed,open")
            .OldAnnotation("Npgsql:Enum:provider_invoice_item_type", "discount,extra_detail,extra_header,extra_location,other,plan,service,usage")
            .OldAnnotation("Npgsql:Enum:provider_invoice_status", "cancelled,draft,overdue,paid,pending")
            .OldAnnotation("Npgsql:Enum:service_application_type", "addon,plan,service")
            .OldAnnotation("Npgsql:Enum:service_availability_rule", "always,custom,cycle_only")
            .OldAnnotation("Npgsql:Enum:service_type", "data,other,roaming,sms,subscription")
            .OldAnnotation("Npgsql:Enum:transition_sub_status", "none,pending_activation,pending_cancellation,pending_portability,pending_pp,pending_tt");

        migrationBuilder.AddColumn<int>(
            name: "ReferenceMonth",
            table: "ProviderInvoices",
            type: "integer",
            nullable: false,
            defaultValue: 0);

        migrationBuilder.AddColumn<int>(
            name: "ReferenceYear",
            table: "ProviderInvoices",
            type: "integer",
            nullable: false,
            defaultValue: 0);

        migrationBuilder.CreateIndex(
            name: "IX_ProviderInvoices_ProviderAccountId_ContractingCompanyId_Bil~",
            table: "ProviderInvoices",
            columns: new[] { "ProviderAccountId", "ContractingCompanyId", "BillingCycleId", "DueDate" },
            unique: true);
    }
}
