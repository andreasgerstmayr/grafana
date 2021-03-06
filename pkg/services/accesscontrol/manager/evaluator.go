package manager

import (
	"context"

	"github.com/gobwas/glob"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
)

const roleGrafanaAdmin = "Grafana Admin"

func (m *Manager) Evaluate(ctx context.Context, user *models.SignedInUser, permission string, scope ...string) (bool, error) {
	roles := []string{string(user.OrgRole)}
	for _, role := range user.OrgRole.Children() {
		roles = append(roles, string(role))
	}
	if user.IsGrafanaAdmin {
		roles = append(roles, roleGrafanaAdmin)
	}

	res, err := m.GetUserPermissions(ctx, accesscontrol.GetUserPermissionsQuery{
		OrgID:  user.OrgId,
		UserID: user.UserId,
		Roles:  roles,
	})
	if err != nil {
		return false, err
	}

	ok, dbScopes := extractPermission(res, permission)
	if !ok {
		return false, nil
	}

	for _, s := range scope {
		var match bool
		for dbScope := range dbScopes {
			rule, err := glob.Compile(dbScope, ':', '/')
			if err != nil {
				return false, err
			}

			match = rule.Match(s)
			if match {
				break
			}
		}

		if !match {
			return false, nil
		}
	}

	return true, nil
}

func extractPermission(permissions []*accesscontrol.Permission, permission string) (bool, map[string]struct{}) {
	scopes := map[string]struct{}{}
	ok := false

	for _, p := range permissions {
		if p == nil {
			continue
		}
		if p.Permission == permission {
			ok = true
			scopes[p.Scope] = struct{}{}
		}
	}

	return ok, scopes
}
