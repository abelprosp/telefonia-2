package auth

import (
	"context"
	"strings"
)

const (
	RoleAdmin     = "admin"
	RoleMaster    = "master"
	RoleEmployee  = "employee"
	RoleOperator  = "operator"
	RoleFinancial = "financial"
	RoleSales     = "sales"
	RoleViewer    = "viewer"
	RolePartner   = "partner"
	RoleUser      = "user"
)

func hasAnyRole(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if HasRole(ctx, role) {
			return true
		}
	}
	return false
}

func IsMaster(ctx context.Context) bool {
	return hasAnyRole(ctx, RoleMaster, RoleAdmin)
}

func IsEmployee(ctx context.Context) bool {
	return hasAnyRole(ctx, RoleEmployee, RoleOperator)
}

func IsFinancial(ctx context.Context) bool {
	return HasRole(ctx, RoleFinancial)
}

func IsSales(ctx context.Context) bool {
	return HasRole(ctx, RoleSales)
}

func IsViewer(ctx context.Context) bool {
	return HasRole(ctx, RoleViewer)
}

func IsInternalStaff(ctx context.Context) bool {
	return IsMaster(ctx) || IsEmployee(ctx) || IsFinancial(ctx) || IsSales(ctx) || IsViewer(ctx)
}

func CanAccessOperational(ctx context.Context) bool {
	return IsMaster(ctx) || IsEmployee(ctx) || IsSales(ctx)
}

func CanAccessFinancial(ctx context.Context) bool {
	return IsMaster(ctx) || IsFinancial(ctx)
}

func CanAccessSales(ctx context.Context) bool {
	return IsMaster(ctx) || IsSales(ctx) || IsEmployee(ctx)
}

func CanMutateOperational(ctx context.Context) bool {
	return IsMaster(ctx) || IsEmployee(ctx)
}

func AllStaffProfiles() []string {
	return []string{RoleMaster, RoleEmployee, RoleOperator, RoleFinancial, RoleSales, RoleViewer, RolePartner, RoleUser}
}

func CanManageUsers(ctx context.Context) bool {
	return IsMaster(ctx)
}

func CanApproveOperations(ctx context.Context) bool {
	return IsMaster(ctx)
}

// CanApproveTwoLevel indica papéis privilegiados que podem participar da aprovação em dois níveis.
func CanApproveTwoLevel(ctx context.Context) bool {
	return IsMaster(ctx) || IsFinancial(ctx)
}

func IsPrivilegedApproverRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case RoleMaster, RoleAdmin, RoleFinancial:
		return true
	default:
		return false
	}
}

func CanAnonymizeData(ctx context.Context) bool {
	return IsMaster(ctx)
}
