package model

type Role string

const (
	RoleLeader      Role = "leader"
	RoleMaintainer  Role = "maintainer"
	RoleCommitter   Role = "committer"
	RoleReviewer    Role = "reviewer"
	RoleContributor Role = "contributor"
)

var (
	ValidRoles = []Role{
		RoleLeader,
		RoleMaintainer,
		RoleCommitter,
		RoleReviewer,
		RoleContributor,
	}
)

func (r Role) String() string {
	return string(r)
}
