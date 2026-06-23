package auth

import "context"

const (
	RoleAdmin     = "admin"
	RoleMaster    = "master"
	RoleEmployee  = "employee"
	RoleFinancial = "financial"
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
	return HasRole(ctx, RoleEmployee)
}

func IsFinancial(ctx context.Context) bool {
	return HasRole(ctx, RoleFinancial)
}

func IsInternalStaff(ctx context.Context) bool {
	return IsMaster(ctx) || IsEmployee(ctx) || IsFinancial(ctx)
}

func CanAccessOperational(ctx context.Context) bool {
	return IsMaster(ctx) || IsEmployee(ctx)
}

func CanAccessFinancial(ctx context.Context) bool {
	return IsMaster(ctx) || IsFinancial(ctx)
}

func CanManageUsers(ctx context.Context) bool {
	return IsMaster(ctx)
}
