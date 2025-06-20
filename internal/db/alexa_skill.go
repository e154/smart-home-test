// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2023, Filippov Alex
//
// This library is free software: you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Library General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with this library.  If not, see
// <https://www.gnu.org/licenses/>.

package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/e154/smart-home/pkg/apperr"
	pkgCommon "github.com/e154/smart-home/pkg/common"
	"gorm.io/gorm"
)

// AlexaSkills ...
type AlexaSkills struct {
	*Common
}

// AlexaSkill ...
type AlexaSkill struct {
	Id          int64 `gorm:"primary_key"`
	SkillId     string
	Description string
	Intents     []*AlexaIntent `gorm:"foreignkey:AlexaSkillId"`
	Status      pkgCommon.StatusType
	Script      *Script
	ScriptId    *int64
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
}

// TableName ...
func (d *AlexaSkill) TableName() string {
	return "alexa_skills"
}

// Add ...
func (n AlexaSkills) Add(ctx context.Context, v *AlexaSkill) (id int64, err error) {
	if err = n.DB(ctx).Create(&v).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillAdd)
		return
	}
	id = v.Id
	return
}

// GetById ...
func (n AlexaSkills) GetById(ctx context.Context, id int64) (v *AlexaSkill, err error) {
	v = &AlexaSkill{Id: id}
	err = n.DB(ctx).Model(v).
		Preload("Script").
		Preload("Intents").
		Preload("Intents.Script").
		Find(v).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("%s: %w", fmt.Sprintf("id \"%d\"", id), apperr.ErrAlexaSkillNotFound)
			return
		}
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillGet)
		return
	}
	if err = n.preload(v); err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillGet)
	}

	return
}

// List ...
func (n *AlexaSkills) List(ctx context.Context, limit, offset int, orderBy, sort string) (list []*AlexaSkill, total int64, err error) {

	if err = n.DB(ctx).Model(AlexaSkill{}).Count(&total).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillList)
		return
	}

	list = make([]*AlexaSkill, 0)
	q := n.DB(ctx).Model(&AlexaSkill{}).
		Limit(limit).
		Offset(offset)

	if sort != "" && orderBy != "" {
		q = q.
			Order(fmt.Sprintf("%s %s", sort, orderBy))
	}

	if err = q.Find(&list).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillList)
	}

	return
}

// ListEnabled ...
func (n *AlexaSkills) ListEnabled(ctx context.Context, limit, offset int) (list []*AlexaSkill, err error) {

	list = make([]*AlexaSkill, 0)
	err = n.DB(ctx).Model(&AlexaSkill{}).
		Where("status = 'enabled'").
		Limit(limit).
		Offset(offset).
		Preload("Intents").
		Preload("Intents.Script").
		Preload("Script").
		Find(&list).Error

	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillList)
		return
	}

	//????
	for _, skill := range list {
		_ = n.preload(skill)
	}

	return
}

func (n AlexaSkills) preload(v *AlexaSkill) (err error) {
	//todo fix
	//err = n.DB(ctx).Model(v).
	//	Related(&v.Intents).Error
	//
	//if err != nil {
	//	err = fmt.Errorf("%s: %w","get related intents failed", )(err,)
	//	return
	//}
	//
	//for _, intent := range v.Intents {
	//	intent.Script = &Script{Id: intent.ScriptId}
	//	err = n.DB(ctx).Model(intent).
	//		Related(intent.Script).Error
	//}
	return
}

// Update ...
func (n AlexaSkills) Update(ctx context.Context, v *AlexaSkill) (err error) {
	q := map[string]interface{}{
		"skill_id":    v.SkillId,
		"status":      v.Status,
		"description": v.Description,
	}
	if v.ScriptId != nil {
		q["script_id"] = pkgCommon.Int64Value(v.ScriptId)
	}
	if err = n.DB(ctx).Model(&AlexaSkill{}).Where("id = ?", v.Id).Updates(q).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillUpdate)
	}
	return
}

// Delete ...
func (n AlexaSkills) Delete(ctx context.Context, id int64) (err error) {
	if err = n.DB(ctx).Delete(&AlexaSkill{}, "id = ?", id).Error; err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrAlexaSkillDelete)
	}
	return
}
