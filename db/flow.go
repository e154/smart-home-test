package db

import (
	"time"
	"fmt"
	"github.com/jinzhu/gorm"
	. "github.com/e154/smart-home/common"
)

type Flows struct {
	Db *gorm.DB
}

type Flow struct {
	Id                 int64 `gorm:"primary_key"`
	Name               string
	Description        string
	Status             StatusType
	WorkflowId         int64
	WorkflowScenarioId int64
	Connections        []*Connection
	FlowElements       []*FlowElement
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (d *Flow) TableName() string {
	return "flows"
}

func (n Flows) Add(flow *Flow) (id int64, err error) {
	if err = n.Db.Create(&flow).Error; err != nil {
		return
	}
	id = flow.Id

	err = n.DependencyLoading(flow)
	return
}

func (n Flows) GetAllEnabled() (list []*Flow, err error) {
	list = make([]*Flow, 0)
	err = n.Db.Where("status = ?", "enabled").
		Find(&list).Error
	if err != nil {
		return
	}

	for _, flow := range list {
		if err = n.DependencyLoading(flow); err != nil {
			return
		}
	}
	return
}

func (n Flows) GetById(flowId int64) (flow *Flow, err error) {
	flow = &Flow{Id: flowId}
	if err = n.Db.First(&flow).Error; err != nil {
		return
	}

	err = n.DependencyLoading(flow)
	return
}

func (n Flows) Update(m *Flow) (err error) {
	err = n.Db.Model(&Flow{Id: m.Id}).Updates(map[string]interface{}{
		"name":        m.Name,
		"description": m.Description,
		"status":      m.Status,
		"workflow_id": m.WorkflowId,
		"scenario_id": m.WorkflowScenarioId,
	}).Error
	return
}

func (n Flows) Delete(flowId int64) (err error) {
	err = n.Db.Delete(&Flow{Id: flowId}).Error
	return
}

func (n *Flows) List(limit, offset int64, orderBy, sort string) (list []*Flow, total int64, err error) {

	if err = n.Db.Model(Flow{}).Count(&total).Error; err != nil {
		return
	}

	list = make([]*Flow, 0)
	err = n.Db.
		Limit(limit).
		Offset(offset).
		Order(fmt.Sprintf("%s %s", sort, orderBy)).
		Find(&list).
		Error
	if err != nil {
		return
	}

	for _, flow := range list {
		if err = n.DependencyLoading(flow); err != nil {
			return
		}
	}
	return
}

func (n *Flows) DependencyLoading(flow *Flow) (err error) {
	flow.Connections = make([]*Connection, 0)
	flow.FlowElements = make([]*FlowElement, 0)

	n.Db.Model(flow).
		Related(&flow.Connections).
		Related(&flow.FlowElements)

	// scripts
	var scriptIds []int64
	for _, element := range flow.FlowElements {
		if element.ScriptId != nil {
			scriptIds = append(scriptIds, *element.ScriptId)
		}
	}

	scripts := make([]*Script, 0)
	err = n.Db.Model(&Script{}).
		Where("id in (?)", scriptIds).
		Find(&scripts).
		Error
	if err != nil {
		return
	}

	for _, element := range flow.FlowElements {
		if element.ScriptId != nil {
			for _, script := range scripts {
				if *element.ScriptId == script.Id {
					element.Script = script
				}
			}
		}
	}

	return
}