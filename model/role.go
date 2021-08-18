package model

type Role string

const (
	RoleAdmin       Role = "admin"
	RoleMaintainer  Role = "maintainer"
	RoleCommitter   Role = "committer"
	RoleReviewer    Role = "reviewer"
	RoleContributor Role = "contributor"
)

func (r Role) String() string {
	return string(r)
}
