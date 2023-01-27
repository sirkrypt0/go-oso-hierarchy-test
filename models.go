package main

type User struct {
	ID   int
	Name string `gorm:"not null;"`

	Teams []UserTeamRole `gorm:"constraint:OnDelete:CASCADE;"`
}

type UserTeamRole struct {
	ID int

	User   User
	UserID int

	Team   Team
	TeamID int

	Role string
}

type Team struct {
	ID   int    `gorm:"not null;type:UUID;"`
	Name string `gorm:"not null;"`

	// Team hierarchy
	Parent   *Team
	ParentID int    `gorm:"default:NULL;"`
	Subteams []Team `gorm:"foreignKey:ParentID;"`

	// Members
	Users []UserTeamRole `gorm:"constraint:OnDelete:CASCADE;"`
}
