using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace Luxus.Connect.Infra.Data.Migrations;

/// <inheritdoc />
public partial class _00 : Migration
{
    /// <inheritdoc />
    protected override void Up(MigrationBuilder migrationBuilder)
    {
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
            .Annotation("Npgsql:Enum:transition_sub_status", "none,pending_activation,pending_cancellation,pending_portability,pending_pp,pending_tt");
    }

    /// <inheritdoc />
    protected override void Down(MigrationBuilder migrationBuilder)
    {

    }
}
