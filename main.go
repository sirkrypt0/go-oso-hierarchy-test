package main

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/osohq/go-oso"
	osoTypes "github.com/osohq/go-oso/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const dbFile = "test.sqlite"

var permissions = []string{"read", "edit", "admin"}

func main() {
	db := setupDB()
	oso := setupOso(db)

	user := User{ID: 1}
	if err := db.Preload(clause.Associations).Find(&user).Error; err != nil {
		log.Fatalf("loading team: %v", err)
	}
	log.Printf("User:\n%#v", user)

	fmt.Println("\n### Checking rootTeam")
	rootTeam := Team{ID: 1}
	if err := db.Preload(clause.Associations).Find(&rootTeam).Error; err != nil {
		log.Fatalf("loading team: %v", err)
	}
	log.Printf("Root Team:\n%#v", rootTeam)

	for _, perm := range permissions {
		if err := oso.Authorize(user, perm, rootTeam); err != nil {
			log.Printf("User cannot %s root team: %v", perm, err)
		} else {
			log.Printf("User %s rootTeam ✔️", perm)
		}
	}

	fmt.Println("\n### Checking sub team")
	subTeam := Team{ID: 2}
	if err := db.Preload(clause.Associations).Find(&subTeam).Error; err != nil {
		log.Fatalf("loading team: %v", err)
	}
	log.Printf("Sub Team:\n%#v", rootTeam)

	for _, perm := range permissions {
		if err := oso.Authorize(user, perm, subTeam); err != nil {
			log.Fatalf("User cannot %s subTeam: %v", perm, err)
		} else {
			log.Printf("User %s subTeam ✔️", perm)
		}
	}
}

func setupOso(db *gorm.DB) *oso.Oso {
	oso, err := oso.NewOso()
	if err != nil {
		log.Fatalf("creating new Oso: %v", err)
	}

	oso.SetDataFilteringAdapter(GormAdapter{
		db:  db,
		oso: &oso,
	})

	oso.RegisterClassWithNameAndFields(reflect.TypeOf(User{}), nil, "User", map[string]interface{}{
		"ID": "Integer",
		"Teams": osoTypes.Relation{
			Kind:       "many",
			OtherType:  "UserTeamRole",
			MyField:    "ID",
			OtherField: "UserID",
		},
	})
	oso.RegisterClassWithNameAndFields(reflect.TypeOf(Team{}), nil, "Team", map[string]interface{}{
		"ID": "Integer",
		"Parent": osoTypes.Relation{
			Kind:       "one",
			OtherType:  "Team",
			MyField:    "ParentID",
			OtherField: "ID",
		},
		"ParentID": "Integer",
	})
	oso.RegisterClassWithNameAndFields(reflect.TypeOf(UserTeamRole{}), nil, "UserTeamRole", map[string]interface{}{
		"ID":     "Integer",
		"UserID": "Integer",
		"TeamID": "Integer",
		"Role":   "String",
	})

	if err := oso.LoadFiles([]string{"main.polar"}); err != nil {
		log.Fatalf("error loading polar file: %v", err)
	}
	return &oso
}

func setupDB() *gorm.DB {
	os.Remove(dbFile)

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&User{}, &UserTeamRole{}, &Team{}); err != nil {
		log.Fatalf("migration: %v", err)
	}

	// Create User
	user := User{ID: 1, Name: "Admin"}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("creating user: %v", err)
	}

	// Create Teams
	rootTeam := Team{ID: 1, Name: "Root"}
	if err := db.Create(&rootTeam).Error; err != nil {
		log.Fatalf("creating root team: %v", err)
	}
	subTeam := Team{ID: 2, Name: "Sub", ParentID: rootTeam.ID}
	if err := db.Create(&subTeam).Error; err != nil {
		log.Fatalf("creating sub team: %v", err)
	}

	// Assign user to teams
	rootUserRole := UserTeamRole{ID: 1, UserID: user.ID, TeamID: rootTeam.ID, Role: "owner"}
	if err := db.Create(&rootUserRole).Error; err != nil {
		log.Fatalf("assigning user to root team: %v", err)
	}

	return db
}