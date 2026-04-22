using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations;

/// <inheritdoc />
public partial class _02 : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.AlterColumn<string>(
            name: "CostCenterId",
            table: "ProviderInvoices",
            type: "character varying(36)",
            maxLength: 36,
            nullable: true,
            oldClrType: typeof(string),
            oldType: "character varying(36)",
            oldMaxLength: 36);
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {
        migrationBuilder.AlterColumn<string>(
            name: "CostCenterId",
            table: "ProviderInvoices",
            type: "character varying(36)",
            maxLength: 36,
            nullable: false,
            defaultValue: "",
            oldClrType: typeof(string),
            oldType: "character varying(36)",
            oldMaxLength: 36,
            oldNullable: true);
    }
}
