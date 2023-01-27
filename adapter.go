package main

import (
	"strings"

	"github.com/osohq/go-oso"
	osoTypes "github.com/osohq/go-oso/types"
	"gorm.io/gorm"
)

// GormAdapter is an adapter for Oso based on https://github.com/osohq/oso/blob/main/languages/go/tests/data_filtering_test.go
type GormAdapter struct {
	db  *gorm.DB
	oso *oso.Oso
}

func (a GormAdapter) sqlize(fc osoTypes.FilterCondition) (string, []interface{}) {
	args := []interface{}{}
	lhs := a.addSide(fc.Lhs, &args)
	rhs := a.addSide(fc.Rhs, &args)
	return lhs + " " + op(fc.Cmp) + " " + rhs, args
}

func op(c osoTypes.Comparison) string {
	switch c {
	case osoTypes.Eq:
		return "="
	case osoTypes.Neq:
		return "!="
	}
	return "IN"
}

func (a GormAdapter) addSide(d osoTypes.Datum, xs *[]interface{}) string {
	switch v := d.DatumVariant.(type) {
	case osoTypes.Projection:
		var fieldName string
		if v.FieldName == "" { // is this how none is returned to Go??
			fieldName = "ID"
		} else {
			fieldName = v.FieldName
		}
		tableName := a.tableName(v.TypeName)
		columnName := a.columnName(tableName, fieldName)
		return tableName + "." + columnName
	case osoTypes.Immediate:
		// not the best way to do this ...
		switch vv := v.Value.(type) {
		case User:
			*xs = append(*xs, vv.ID)
		case Team:
			*xs = append(*xs, vv.ID)
		case UserTeamRole:
			*xs = append(*xs, vv.ID)
		default:
			*xs = append(*xs, vv)
		}
	}
	return "?"
}

func (a GormAdapter) tableName(n string) string {
	return a.db.Config.NamingStrategy.TableName(n)
}

func (a GormAdapter) columnName(t string, n string) string {
	return a.db.Config.NamingStrategy.ColumnName(t, n)
}

func (a GormAdapter) BuildQuery(f *osoTypes.Filter) (interface{}, error) {
	models := map[string]interface{}{
		"User":         User{},
		"Team":         Team{},
		"UserTeamRole": UserTeamRole{},
	}
	model := models[f.Root]
	db := a.db.Table(a.tableName(f.Root)).Model(&model)

	for _, rel := range f.Relations {
		myTable := a.tableName(rel.FromTypeName)
		otherTable := a.tableName(rel.ToTypeName)
		myField, otherField, err := a.oso.GetHost().GetRelationFields(rel)
		if err != nil {
			return nil, err
		}
		myColumn := a.columnName(myTable, myField)
		otherColumn := a.columnName(otherTable, otherField)
		join := "INNER JOIN " + otherTable + " ON " + myTable + "." + myColumn + " = " + otherTable + "." + otherColumn
		db = db.Joins(join)
	}

	orSqls := []string{}
	args := []interface{}{}
	for _, orClause := range f.Conditions {
		andSqls := []string{}
		for _, andClause := range orClause {
			andSql, andArgs := a.sqlize(andClause)
			andSqls = append(andSqls, andSql)
			args = append(args, andArgs...)
		}

		if len(andSqls) > 0 {
			orSqls = append(orSqls, strings.Join(andSqls, " AND "))
		}
	}

	if len(orSqls) > 0 {
		sql := strings.Join(orSqls, " OR ")
		db = db.Where(sql, args...)
	}

	return db, nil
}

func (a GormAdapter) ExecuteQuery(x interface{}) ([]interface{}, error) {
	switch q := x.(type) {
	case *gorm.DB:
		switch (*q.Statement.Model.(*interface{})).(type) {
		case User:
			v := make([]User, 0)
			q.Distinct().Find(&v)
			w := make([]interface{}, 0)
			for _, i := range v {
				w = append(w, i)
			}
			// log.Printf("ExecuteQuery User: %#v", w)
			return w, nil
		case Team:
			v := make([]Team, 0)
			q.Distinct().Find(&v)
			w := make([]interface{}, 0)
			for _, i := range v {
				w = append(w, i)
			}
			// log.Printf("ExecuteQuery Team: %#v", w)
			return w, nil
		case UserTeamRole:
			v := make([]UserTeamRole, 0)
			q.Distinct().Find(&v)
			w := make([]interface{}, 0)
			for _, i := range v {
				w = append(w, i)
			}
			// log.Printf("ExecuteQuery UserTeamRole: %#v", w)
			return w, nil
		}
	}
	panic("a problem")
}
