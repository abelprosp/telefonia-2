using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations;

/// <inheritdoc />
public partial class _03 : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropForeignKey(
            name: "FK_ProviderInvoices_ProviderInvoices_ParentInvoiceId",
            table: "ProviderInvoices");

        migrationBuilder.AlterColumn<string>(
            name: "ParentInvoiceId",
            table: "ProviderInvoices",
            type: "character varying(36)",
            maxLength: 36,
            nullable: true,
            oldClrType: typeof(string),
            oldType: "character varying(36)",
            oldMaxLength: 36);

        migrationBuilder.AddForeignKey(
            name: "FK_ProviderInvoices_ProviderInvoices_ParentInvoiceId",
            table: "ProviderInvoices",
            column: "ParentInvoiceId",
            principalTable: "ProviderInvoices",
            principalColumn: "Id");
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.DropForeignKey(
            name: "FK_ProviderInvoices_ProviderInvoices_ParentInvoiceId",
            table: "ProviderInvoices");

        migrationBuilder.AlterColumn<string>(
            name: "ParentInvoiceId",
            table: "ProviderInvoices",
            type: "character varying(36)",
            maxLength: 36,
            nullable: false,
            defaultValue: "",
            oldClrType: typeof(string),
            oldType: "character varying(36)",
            oldMaxLength: 36,
            oldNullable: true);

        migrationBuilder.AddForeignKey(
            name: "FK_ProviderInvoices_ProviderInvoices_ParentInvoiceId",
            table: "ProviderInvoices",
            column: "ParentInvoiceId",
            principalTable: "ProviderInvoices",
            principalColumn: "Id",
            onDelete: ReferentialAction.Cascade);
    }
}
