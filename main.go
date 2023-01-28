package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/osohq/go-oso"
	osoTypes "github.com/osohq/go-oso/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const dbFile = "test.sqlite"

var (
	userId      = uuid.New()
	rootId      = TeamID{uuid.New()}
	subId       = TeamID{uuid.New()}
	subSubId    = TeamID{uuid.New()}
	otherId     = TeamID{uuid.New()}
	repoId      = uuid.New()
	otherRepoId = uuid.New()
)

func main() {
	db := setupDB()
	oso := setupOso(db)

	user := User{ID: userId}
	if err := db.Preload(clause.Associations).Find(&user).Error; err != nil {
		log.Fatalf("loading team: %v", err)
	}

	teams := []Team{
		{ID: rootId},
		{ID: subId},
		{ID: subSubId},
		{ID: otherId},
	}

	for _, t := range teams {
		if err := db.Preload(clause.Associations).Find(&t).Error; err != nil {
			log.Fatalf("loading team: %v", err)
		}
		fmt.Printf("%s: ", t.Name)

		actions, err := oso.AuthorizedActions(user, t, true)
		if err != nil {
			log.Fatalf("Couldn't get authorized actions for user: %v", err)
		}
		permissions := []string{}
		for perm := range actions {
			permissions = append(permissions, perm.(string))
		}
		fmt.Printf("%v\n", permissions)
	}

	fmt.Println("\n### Checking repository of root team")
	repository := Repository{ID: repoId}
	if err := db.Preload(clause.Associations).Find(&repository).Error; err != nil {
		log.Fatalf("loading repository: %v", err)
	}

	// AuthorizedResources for resources that have a relation to a self-relating
	// resource doesn't work for now.
	// See https://docs.osohq.com/go/guides/data_filtering.html

	// resources, err := oso.AuthorizedResources(user, "read", "Repository")
	// if err != nil {
	// 	log.Fatalf("Couldn't get authorized resources for user: %v", err)
	// }
	// fmt.Printf("Authorized Resources: %#v", resources)

	actions, err := oso.AuthorizedActions(user, repository, true)
	if err != nil {
		log.Fatalf("Couldn't get authorized actions for user: %v", err)
	}
	permissions := []string{}
	for perm := range actions {
		permissions = append(permissions, perm.(string))
	}
	fmt.Printf("Authorized actions: %v\n", permissions)
}

func setupOso(db *gorm.DB) *oso.Oso {
	oso, err := oso.NewOso()
	if err != nil {
		log.Fatalf("creating new Oso: %v", err)
	}

	oso.SetDataFilteringAdapter(GormAdapter{
		db:    db,
		oso:   &oso,
		debug: false,
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
	// oso.RegisterClassWithNameAndFields(reflect.TypeOf(UserTeamRole{}), nil, "UserTeamRole", map[string]interface{}{
	// 	"UserID": "Integer",
	// 	"TeamID": "Integer",
	// 	"Role":   "String",
	// })

	oso.RegisterClassWithNameAndFields(reflect.TypeOf(Team{}), nil, "Team", map[string]interface{}{
		"ID": "TeamID",
		"Parent": osoTypes.Relation{
			Kind:       "one",
			OtherType:  "Team",
			MyField:    "ParentID",
			OtherField: "ID",
		},
		"ParentID": "TeamID",
	})

	oso.RegisterClassWithNameAndFields(reflect.TypeOf(Repository{}), nil, "Repository", map[string]interface{}{
		"Team": osoTypes.Relation{
			Kind:       "one",
			OtherType:  "Team",
			MyField:    "TeamID",
			OtherField: "ID",
		},
		"TeamID": "TeamID",
	})

	if err := oso.LoadFiles([]string{"main.polar"}); err != nil {
		log.Fatalf("error loading polar file: %v", err)
	}
	return &oso
}

func setupDB() *gorm.DB {
	os.Remove(dbFile)

	dbLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      true,
	})

	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&User{}, &UserTeamRole{}, &Team{}, &Repository{}); err != nil {
		log.Fatalf("migration: %v", err)
	}

	objects := []interface{}{
		// User
		User{ID: userId, Name: "Admin"},
		// Teams
		Team{ID: rootId, Name: "Root"},
		Team{ID: subId, Name: "Sub", ParentID: &rootId},
		Team{ID: subSubId, Name: "SubSub", ParentID: &subId},
		Team{ID: otherId, Name: "OtherTeam"},
		// User Team Assignment
		UserTeamRole{UserID: userId, TeamID: rootId, Role: "guest"},
		UserTeamRole{UserID: userId, TeamID: subId, Role: "owner"},
		UserTeamRole{UserID: userId, TeamID: otherId, Role: "guest"},
		// Repos
		Repository{ID: repoId, TeamID: rootId},
		Repository{ID: otherRepoId, TeamID: otherId},
	}

	for _, o := range objects {
		var err error
		switch o := o.(type) {
		case User:
			err = db.Create(&o).Error
		case Team:
			err = db.Create(&o).Error
		case UserTeamRole:
			err = db.Create(&o).Error
		case Repository:
			err = db.Create(&o).Error
		}
		if err != nil {
			log.Fatalf("creating objects: %v", err)
		}
	}

	return db
}
