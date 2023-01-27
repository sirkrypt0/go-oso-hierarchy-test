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
	log.Printf("Sub Team:\n%#v", subTeam)

	for _, perm := range permissions {
		if err := oso.Authorize(user, perm, subTeam); err != nil {
			log.Fatalf("User cannot %s subTeam: %v", perm, err)
		} else {
			log.Printf("User %s subTeam ✔️", perm)
		}
	}

	fmt.Println("\n### Checking sub sub team")
	subSubTeam := Team{ID: 3}
	if err := db.Preload(clause.Associations).Find(&subSubTeam).Error; err != nil {
		log.Fatalf("loading subsub team: %v", err)
	}
	log.Printf("SubSub Team:\n%#v", subSubTeam)
	log.Printf("SubSub Team Parent:\n%#v", subSubTeam.Parent)

	for _, perm := range permissions {
		if err := oso.Authorize(user, perm, subSubTeam); err != nil {
			log.Fatalf("User cannot %s subSubTeam: %v", perm, err)
		} else {
			log.Printf("User %s subSubTeam ✔️", perm)
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

	// If we make sure that the user's teams are preloaded, we don't need to specify the teams relation here.
	// Otherwise, Oso would query the teams itself every time it evaluates a `role in user.Teams` rule.
	oso.RegisterClass(reflect.TypeOf(User{}), nil)

	// oso.RegisterClassWithNameAndFields(reflect.TypeOf(User{}), nil, "User", map[string]interface{}{
	// 	"ID": "Integer",
	// 	"Teams": osoTypes.Relation{
	// 		Kind:       "many",
	// 		OtherType:  "UserTeamRole",
	// 		MyField:    "ID",
	// 		OtherField: "UserID",
	// 	},
	// })

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
	subSubTeam := Team{ID: 3, Name: "SubSub", ParentID: subTeam.ID}
	if err := db.Create(&subSubTeam).Error; err != nil {
		log.Fatalf("creating sub sub team: %v", err)
	}
	otherTeam := Team{ID: 10, Name: "OtherTeam"}
	if err := db.Create(&otherTeam).Error; err != nil {
		log.Fatalf("creating other team: %v", err)
	}

	// Assign user to teams
	rootUserRole := UserTeamRole{ID: 1, UserID: user.ID, TeamID: rootTeam.ID, Role: "owner"}
	if err := db.Create(&rootUserRole).Error; err != nil {
		log.Fatalf("assigning user to root team: %v", err)
	}
	otherUserRole := UserTeamRole{ID: 2, UserID: user.ID, TeamID: otherTeam.ID, Role: "guest"}
	if err := db.Create(&otherUserRole).Error; err != nil {
		log.Fatalf("assigning user to other team: %v", err)
	}

	return db
}
