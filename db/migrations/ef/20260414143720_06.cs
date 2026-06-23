using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations
{
    /// <inheritdoc />
    public partial class _06 : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.CreateIndex(
                name: "IX_PhoneLines_Number",
                table: "PhoneLines",
                column: "Number",
                unique: true);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropIndex(
                name: "IX_PhoneLines_Number",
                table: "PhoneLines");
        }
    }
}
