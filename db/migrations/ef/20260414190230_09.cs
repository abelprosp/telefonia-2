using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations
{
    /// <inheritdoc />
    public partial class _09 : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropForeignKey(
                name: "FK_Customers_Providers_ProviderId",
                table: "Customers");

            migrationBuilder.DropIndex(
                name: "IX_Customers_OrganizationId_ProviderId_LegalName",
                table: "Customers");

            migrationBuilder.DropIndex(
                name: "IX_Customers_OrganizationId_ProviderId_Name",
                table: "Customers");

            migrationBuilder.DropIndex(
                name: "IX_Customers_ProviderId",
                table: "Customers");

            migrationBuilder.DropColumn(
                name: "ProviderId",
                table: "Customers");

            migrationBuilder.CreateTable(
                name: "CustomerProviderLinks",
                columns: table => new
                {
                    Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                    CustomerId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                    ProviderId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                    StartDate = table.Column<DateOnly>(type: "date", nullable: false),
                    EndDate = table.Column<DateOnly>(type: "date", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_CustomerProviderLinks", x => x.Id);
                    table.ForeignKey(
                        name: "FK_CustomerProviderLinks_Customers_CustomerId",
                        column: x => x.CustomerId,
                        principalTable: "Customers",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                    table.ForeignKey(
                        name: "FK_CustomerProviderLinks_Providers_ProviderId",
                        column: x => x.ProviderId,
                        principalTable: "Providers",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateIndex(
                name: "IX_Customers_OrganizationId_LegalName",
                table: "Customers",
                columns: new[] { "OrganizationId", "LegalName" },
                unique: true);

            migrationBuilder.CreateIndex(
                name: "IX_Customers_OrganizationId_Name",
                table: "Customers",
                columns: new[] { "OrganizationId", "Name" });

            migrationBuilder.CreateIndex(
                name: "IX_CustomerProviderLinks_CustomerId_ProviderId",
                table: "CustomerProviderLinks",
                columns: new[] { "CustomerId", "ProviderId" },
                unique: true,
                filter: "\"EndDate\" IS NULL");

            migrationBuilder.CreateIndex(
                name: "IX_CustomerProviderLinks_CustomerId_ProviderId_EndDate",
                table: "CustomerProviderLinks",
                columns: new[] { "CustomerId", "ProviderId", "EndDate" });

            migrationBuilder.CreateIndex(
                name: "IX_CustomerProviderLinks_CustomerId_ProviderId_StartDate",
                table: "CustomerProviderLinks",
                columns: new[] { "CustomerId", "ProviderId", "StartDate" });

            migrationBuilder.CreateIndex(
                name: "IX_CustomerProviderLinks_ProviderId",
                table: "CustomerProviderLinks",
                column: "ProviderId");
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "CustomerProviderLinks");

            migrationBuilder.DropIndex(
                name: "IX_Customers_OrganizationId_LegalName",
                table: "Customers");

            migrationBuilder.DropIndex(
                name: "IX_Customers_OrganizationId_Name",
                table: "Customers");

            migrationBuilder.AddColumn<string>(
                name: "ProviderId",
                table: "Customers",
                type: "character varying(36)",
                maxLength: 36,
                nullable: false,
                defaultValue: "");

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

            migrationBuilder.AddForeignKey(
                name: "FK_Customers_Providers_ProviderId",
                table: "Customers",
                column: "ProviderId",
                principalTable: "Providers",
                principalColumn: "Id",
                onDelete: ReferentialAction.Cascade);
        }
    }
}
