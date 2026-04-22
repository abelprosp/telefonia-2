using System;
using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations
{
    /// <inheritdoc />
    public partial class _07 : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateTable(
                name: "PhoneLineCustomerLinks",
                columns: table => new
                {
                    Id = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                    PhoneLineId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                    CustomerId = table.Column<string>(type: "character varying(36)", maxLength: 36, nullable: false),
                    StartDate = table.Column<DateOnly>(type: "date", nullable: false),
                    EndDate = table.Column<DateOnly>(type: "date", nullable: true)
                },
                constraints: table =>
                {
                    table.PrimaryKey("PK_PhoneLineCustomerLinks", x => x.Id);
                    table.ForeignKey(
                        name: "FK_PhoneLineCustomerLinks_Customers_CustomerId",
                        column: x => x.CustomerId,
                        principalTable: "Customers",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Restrict);
                    table.ForeignKey(
                        name: "FK_PhoneLineCustomerLinks_PhoneLines_PhoneLineId",
                        column: x => x.PhoneLineId,
                        principalTable: "PhoneLines",
                        principalColumn: "Id",
                        onDelete: ReferentialAction.Cascade);
                });

            migrationBuilder.CreateIndex(
                name: "IX_PhoneLineCustomerLinks_CustomerId",
                table: "PhoneLineCustomerLinks",
                column: "CustomerId");

            migrationBuilder.CreateIndex(
                name: "IX_PhoneLineCustomerLinks_PhoneLineId",
                table: "PhoneLineCustomerLinks",
                column: "PhoneLineId",
                unique: true,
                filter: "\"EndDate\" IS NULL");

            migrationBuilder.CreateIndex(
                name: "IX_PhoneLineCustomerLinks_PhoneLineId_StartDate",
                table: "PhoneLineCustomerLinks",
                columns: new[] { "PhoneLineId", "StartDate" });
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropTable(
                name: "PhoneLineCustomerLinks");
        }
    }
}
