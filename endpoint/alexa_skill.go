// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2021, Filippov Alex
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

package endpoint

import (
	"github.com/e154/smart-home/common"
	m "github.com/e154/smart-home/models"
	"github.com/e154/smart-home/plugins/alexa"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// AlexaSkillEndpoint ...
type AlexaSkillEndpoint struct {
	*CommonEndpoint
}

// NewAlexaSkillEndpoint ...
func NewAlexaSkillEndpoint(common *CommonEndpoint) *AlexaSkillEndpoint {
	return &AlexaSkillEndpoint{
		CommonEndpoint: common,
	}
}

// Add ...
func (n *AlexaSkillEndpoint) Add(params *m.AlexaSkill) (result *m.AlexaSkill, errs validator.ValidationErrorsTranslations, err error) {

	var ok bool
	if ok, errs = n.validation.Valid(params); !ok {
		return
	}

	var id int64
	if id, err = n.adaptors.AlexaSkill.Add(params); err != nil {
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}

	result, err = n.adaptors.AlexaSkill.GetById(id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return
		}
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}

	n.eventBus.Publish(alexa.TopicPluginAlexa, alexa.EventAlexaAddSkill{
		Skill: result,
	})

	return
}

// GetById ...
func (n *AlexaSkillEndpoint) GetById(appId int64) (result *m.AlexaSkill, err error) {

	result, err = n.adaptors.AlexaSkill.GetById(appId)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return
		}
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}
	return
}

// Update ...
func (n *AlexaSkillEndpoint) Update(params *m.AlexaSkill) (skill *m.AlexaSkill, errs validator.ValidationErrorsTranslations, err error) {

	var ok bool
	if ok, errs = n.validation.Valid(params); !ok {
		return
	}

	if err = n.adaptors.AlexaSkill.Update(params); err != nil {
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}

	skill, err = n.adaptors.AlexaSkill.GetById(params.Id)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return
		}
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}

	n.eventBus.Publish(alexa.TopicPluginAlexa, alexa.EventAlexaUpdateSkill{
		Skill: skill,
	})

	return
}

// GetList ...
func (n *AlexaSkillEndpoint) GetList(limit, offset int64, order, sortBy string) (result []*m.AlexaSkill, total int64, err error) {

	result, total, err = n.adaptors.AlexaSkill.List(limit, offset, order, sortBy)
	if err != nil {
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}
	return
}

// Delete ...
func (n *AlexaSkillEndpoint) Delete(skillId int64) (err error) {

	if skillId == 0 {
		err = common.ErrBadRequestParams
		return
	}

	var skill *m.AlexaSkill
	skill, err = n.adaptors.AlexaSkill.GetById(skillId)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return
		}
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}

	if err = n.adaptors.AlexaSkill.Delete(skill.Id); err != nil {
		err = errors.Wrap(common.ErrInternal, err.Error())
		return
	}

	n.eventBus.Publish(alexa.TopicPluginAlexa, alexa.EventAlexaDeleteSkill{
		Skill: skill,
	})

	return
}