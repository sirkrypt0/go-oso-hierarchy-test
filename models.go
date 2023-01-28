package main

import "github.com/google/uuid"

type User struct {
	ID   uuid.UUID
	Name string `gorm:"not null;"`

	Teams []UserTeamRole `gorm:"constraint:OnDelete:CASCADE;"`
}

type UserTeamRole struct {
	ID uuid.UUID

	User   User
	UserID uuid.UUID

	Team   Team
	TeamID TeamID

	Role string
}

// TeamID is the identifier of a team.
// We use a separate type instead of uuid.UUID directly,
// as it allows us to implement Oso's Comparer interface and then
// use the custom uuid.UUID type for = comparison in the Polar definition.
type TeamID struct {
	uuid.UUID
}

func (t TeamID) Equal(other TeamID) bool {
	return t.UUID == other.UUID
}

func (t TeamID) Lt(other TeamID) bool {
	return false
}

type Team struct {
	ID   TeamID `gorm:"not null;"`
	Name string `gorm:"not null;"`

	// Team hierarchy
	Parent   *Team
	ParentID *TeamID `gorm:"default:NULL;"`
	Subteams []Team  `gorm:"foreignKey:ParentID;"`

	// Members
	Users []UserTeamRole `gorm:"constraint:OnDelete:CASCADE;"`
}

type Repository struct {
	ID     uuid.UUID
	Team   Team   `gorm:"constraint:OnDelete:CASCADE;"`
	TeamID TeamID `gorm:"not null;constraint:OnDelete:CASCADE;"`
}
