package gf_casbin

import (
	"context"
	"fmt"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gogf/gf/v2/database/gdb"
	"runtime"
)

// CasbinRule is the golang structure for table casbin_rule.
type CasbinRule struct {
	Ptype string `json:"ptype" ` //
	V0    string `json:"v0"    ` //
	V1    string `json:"v1"    ` //
	V2    string `json:"v2"    ` //
	V3    string `json:"v3"    ` //
	V4    string `json:"v4"    ` //
	V5    string `json:"v5"    ` //
}

// Adapter 用于策略存储的gdb适配器。
type Adapter struct {
	Link      string // 链接到数据库的信息，例如: "mysql:root:12345678@tcp(127.0.0.1:3306)/test"
	TableName string // 存储策略的表
	DB        gdb.DB // 操作数据库的ORM句柄
	ctx       context.Context
}

// finalizer is the destructor for Adapter.
func finalizer(a *Adapter) {
	// 注意不用的时候不需要使用Close方法关闭数据库连接(并且gdb也没有提供Close方法)，
	// 数据库引擎底层采用了链接池设计，当链接不再使用时会自动关闭
	a.DB = nil
}

// NewAdapter is the constructor for Adapter.
func NewAdapter(ctx context.Context, link string) (*Adapter, error) {
	a := &Adapter{
		Link:      link,
		TableName: "casbin_rule",
		ctx:       ctx,
	}
	// Open the DB, create it if not existed.
	if err := a.open(); err != nil {
		return nil, err
	}

	// Call the destructor when the object is released.
	runtime.SetFinalizer(a, finalizer)

	return a, nil
}

func (a *Adapter) open() error {
	gdb.SetConfig(gdb.Config{
		"casbin": gdb.ConfigGroup{
			gdb.ConfigNode{
				Link:   a.Link,
				Role:   "master",
				Weight: 100,
			},
		},
	})

	db, err := gdb.NewByGroup("casbin")
	if err != nil {
		return err
	}
	a.DB = db

	return a.createTable()
}

func (a *Adapter) close() error {
	// 注意不用的时候不需要使用Close方法关闭数据库连接(并且gdb也没有提供Close方法)，
	// 数据库引擎底层采用了链接池设计，当链接不再使用时会自动关闭
	a.DB = nil
	return nil
}

func (a *Adapter) createTable() error {
	_, err := a.DB.Exec(a.ctx, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (ptype VARCHAR(10), v0 VARCHAR(256), v1 VARCHAR(256), v2 VARCHAR(256), v3 VARCHAR(256), v4 VARCHAR(256), v5 VARCHAR(256))", a.TableName))
	return err
}

func (a *Adapter) dropTable() error {
	_, err := a.DB.Exec(a.ctx, fmt.Sprintf("DROP TABLE %s", a.TableName))
	return err
}

func (a *Adapter) DbCtx() *gdb.Model {
	return a.DB.Model(a.TableName).Safe().Ctx(a.ctx)
}

// LoadPolicy loads policy from database.
func (a *Adapter) LoadPolicy(model model.Model) error {
	var lines []*CasbinRule

	if err := a.DbCtx().Scan(&lines); err != nil {
		return err
	}

	for _, line := range lines {
		err := loadPolicyLine(line, model)
		if err != nil {
			return err
		}
	}

	return nil
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model model.Model) error {
	err := a.dropTable()
	if err != nil {
		return err
	}
	err = a.createTable()
	if err != nil {
		return err
	}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			_, err := a.DbCtx().Data(&line).Insert()
			if err != nil {
				return err
			}
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			_, err := a.DbCtx().Data(&line).Insert()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	_, err := a.DbCtx().Data(&line).Insert()
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)
	err := rawDelete(a, line)
	return err
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
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
	err := rawDelete(a, line)
	return err
}

func loadPolicyLine(line *CasbinRule, model model.Model) error {
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

	return persist.LoadPolicyLine(lineText, model)
}

func savePolicyLine(ptype string, rule []string) *CasbinRule {
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

func rawDelete(a *Adapter, line *CasbinRule) error {

	m := a.DbCtx().Where("ptype = ?", line.Ptype)
	if line.V0 != "" {
		m = m.Where("v0 = ?", line.V0)
	}
	if line.V1 != "" {
		m = m.Where("v1 = ?", line.V1)
	}
	if line.V2 != "" {
		m = m.Where("v2 = ?", line.V2)
	}
	if line.V3 != "" {
		m = m.Where("v3 = ?", line.V3)
	}
	if line.V4 != "" {
		m = m.Where("v4 = ?", line.V4)
	}
	if line.V5 != "" {
		m = m.Where("v5 = ?", line.V5)
	}

	_, err := m.Delete()
	return err
}
