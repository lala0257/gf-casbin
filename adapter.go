package gf_casbin

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gogf/gf/v2/database/gdb"
)

const DefaultTableName = "casbin_rule"

type CasbinRule struct {
	Ptype string `json:"ptype" ` //
	V0    string `json:"v0"    ` //
	V1    string `json:"v1"    ` //
	V2    string `json:"v2"    ` //
	V3    string `json:"v3"    ` //
	V4    string `json:"v4"    ` //
	V5    string `json:"v5"    ` //
}

type GfCasbin struct {
	Enforcer    *casbin.SyncedEnforcer
	EnforcerErr error
	modelFile   string
	ctx         context.Context
	db          gdb.DB
}

// New casbin适配器
func New(ctx context.Context, modelFile string, db gdb.DB) *GfCasbin {
	a := &GfCasbin{
		modelFile: modelFile,
	}
	a.initPolicy(ctx, db)
	return a
}

func (a *GfCasbin) DbCtx() *gdb.Model {
	return a.db.Model(DefaultTableName).Safe().Ctx(a.ctx)
}

func (a *GfCasbin) initPolicy(ctx context.Context, db gdb.DB) {
	e, err := casbin.NewSyncedEnforcer(a.modelFile, a)
	if err != nil {
		a.EnforcerErr = err
		return
	}
	e.ClearPolicy()
	a.Enforcer = e
	a.ctx = ctx
	a.db = db
	// Load the policy from DB.
	err = a.LoadPolicy(e.GetModel())
	if err != nil {
		a.EnforcerErr = err
		return
	}
}

// LoadPolicy loads policy from database.
func (a *GfCasbin) LoadPolicy(model model.Model) error {
	var lines []*CasbinRule
	if err := a.DbCtx().Scan(&lines); err != nil {
		return err
	}
	for _, line := range lines {
		a.loadPolicyLine(line, model)
	}
	return nil
}

// SavePolicy saves policy to database.
func (a *GfCasbin) SavePolicy(model model.Model) (err error) {
	err = a.dropTable()
	if err != nil {
		return
	}
	err = a.createTable()
	if err != nil {
		return
	}
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := a.savePolicyLine(ptype, rule)
			_, err := a.DbCtx().Data(line).Insert()
			if err != nil {
				return err
			}
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := a.savePolicyLine(ptype, rule)
			_, err := a.DbCtx().Data(line).Insert()
			if err != nil {
				return err
			}
		}
	}
	return
}

func (a *GfCasbin) dropTable() (err error) {
	return
}

func (a *GfCasbin) createTable() (err error) {
	return
}

// AddPolicy adds a policy rule to the storage.
func (a *GfCasbin) AddPolicy(sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	_, err := a.DbCtx().Data(line).Insert()
	return err
}

func (a *GfCasbin) RemovePolicy(sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	err := a.rawDelete(line)
	return err
}

func (a *GfCasbin) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	line := &CasbinRule{}
	line.Ptype = ptype
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	err := a.rawDelete(line)
	return err
}

func (a *GfCasbin) loadPolicyLine(line *CasbinRule, model model.Model) {
	lineText := line.Ptype
	if line.V0 != "" {
		lineText += ", " + line.V0
	}
	if line.V1 != "" {
		lineText += ", " + line.V1
	}
	if line.V2 != "" {
		lineText += ", " + line.V2
	}
	if line.V3 != "" {
		lineText += ", " + line.V3
	}
	if line.V4 != "" {
		lineText += ", " + line.V4
	}
	if line.V5 != "" {
		lineText += ", " + line.V5
	}
	_ = persist.LoadPolicyLine(lineText, model)
}

func (a *GfCasbin) savePolicyLine(ptype string, rule []string) *CasbinRule {
	line := &CasbinRule{}
	line.Ptype = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}
	return line
}

func (a *GfCasbin) rawDelete(line *CasbinRule) error {
	db := a.DbCtx().Where("ptype = ?", line.Ptype)
	if line.V0 != "" {
		db = db.Where("v0 = ?", line.V0)
	}
	if line.V1 != "" {
		db = db.Where("v1 = ?", line.V1)
	}
	if line.V2 != "" {
		db = db.Where("v2 = ?", line.V2)
	}
	if line.V3 != "" {
		db = db.Where("v3 = ?", line.V3)
	}
	if line.V4 != "" {
		db = db.Where("v4 = ?", line.V4)
	}
	if line.V5 != "" {
		db = db.Where("v5 = ?", line.V5)
	}
	_, err := db.Delete()
	return err
}
