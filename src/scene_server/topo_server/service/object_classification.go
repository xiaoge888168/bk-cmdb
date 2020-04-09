/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"strconv"

	"configcenter/src/common/blog"
	"configcenter/src/common/condition"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/scene_server/topo_server/core/operation"
)

// CreateClassification create a new object classification
func (s *Service) CreateClassification(ctx *rest.Contexts) {
	data := make(map[string]interface{})
	if err := ctx.DecodeInto(&data); err != nil {
		ctx.RespAutoError(err)
		return
	}
	cls, err := s.Core.ClassificationOperation().CreateClassification(ctx.Kit, data)
	if nil != err {
		ctx.RespAutoError(err)
		return
	}
	id, err := cls.ToMapStr().Int64("id")
	if err != nil {
		blog.Errorf("create object classification success, but get response id failed, err: %+v, rid: %s", err, ctx.Kit.Rid)
	}
	objClsAuditLog := operation.NewObjectClsAudit(s.Engine.CoreAPI)
	//get CurData
	err = objClsAuditLog.MakeCurrent(cls.ToMapStr())
	if err != nil {
		blog.Errorf("[api-cls] make Current object classification failed, id: %+v, err: %s, rid: %s", id, err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
	}

	//package audit response
	err = objClsAuditLog.SaveAuditLog(ctx.Kit, metadata.AuditCreate)
	if err != nil {
		ctx.RespAutoError(err)
	}

	ctx.RespEntity(cls.ToMapStr())
}

// SearchClassificationWithObjects search the classification with objects
func (s *Service) SearchClassificationWithObjects(ctx *rest.Contexts) {
	dataWithMetadata := MapStrWithMetadata{}
	if err := ctx.DecodeInto(&dataWithMetadata); err != nil {
		ctx.RespAutoError(err)
		return
	}
	data := dataWithMetadata.Data

	cond := condition.CreateCondition()
	if data.Exists(metadata.PageName) {
		page, err := data.MapStr(metadata.PageName)
		if nil != err {
			blog.Errorf("failed to get the page , error info is %s, rid: %s", err.Error(), ctx.Kit.Rid)
			ctx.RespAutoError(err)
			return
		}

		if err = cond.SetPage(page); nil != err {
			blog.Errorf("failed to parse the page, error info is %s, rid: %s", err.Error(), ctx.Kit.Rid)
			ctx.RespAutoError(err)
			return
		}

		data.Remove(metadata.PageName)
	}

	if err := cond.Parse(data); nil != err {
		blog.Errorf("failed to parse the condition, error info is %s, rid: %s", err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	resp, err := s.Core.ClassificationOperation().FindClassificationWithObjects(ctx.Kit, cond, dataWithMetadata.Metadata)
	if err != nil {
		ctx.RespAutoError(err)
		return
	}
	ctx.RespEntity(resp)
}

// SearchClassification search the classifications
func (s *Service) SearchClassification(ctx *rest.Contexts) {
	dataWithMetadata := MapStrWithMetadata{}
	if err := ctx.DecodeInto(&dataWithMetadata); err != nil {
		ctx.RespAutoError(err)
		return
	}
	data := dataWithMetadata.Data

	cond := condition.CreateCondition()
	if data.Exists(metadata.PageName) {

		page, err := data.MapStr(metadata.PageName)
		if nil != err {
			blog.Errorf("failed to get the page , error info is %s, rid: %s", err.Error(), ctx.Kit.Rid)
			ctx.RespAutoError(err)
			return
		}

		if err = cond.SetPage(page); nil != err {
			blog.Errorf("failed to parse the page, error info is %s, rid: %s", err.Error(), ctx.Kit.Rid)
			ctx.RespAutoError(err)
			return
		}

		data.Remove(metadata.PageName)
	}
	if err := cond.Parse(data); err != nil {
		blog.Errorf("parse condition from data failed, err: %s, rid: %s", err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	resp, err := s.Core.ClassificationOperation().FindClassification(ctx.Kit, cond, dataWithMetadata.Metadata)
	if err != nil {
		ctx.RespAutoError(err)
		return
	}
	ctx.RespEntity(resp)
}

// UpdateClassification update the object classification
func (s *Service) UpdateClassification(ctx *rest.Contexts) {
	data := make(mapstr.MapStr)
	if err := ctx.DecodeInto(&data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	cond := condition.CreateCondition()
	paramPath := mapstr.MapStr{}
	paramPath.Set("id", ctx.Request.PathParameter("id"))
	id, err := paramPath.Int64("id")
	if nil != err {
		blog.Errorf("[api-cls] failed to parse the path params id(%s), error info is %s , rid: %s", ctx.Request.PathParameter("id"), err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}
	data.Remove(metadata.BKMetadata)

	objClsAuditLog := operation.NewObjectClsAudit(s.Engine.CoreAPI)
	//get AuditLog PreData
	err = objClsAuditLog.WithPrevious(ctx.Kit, id)
	if err != nil {
		blog.Errorf("[api-cls] find Previous objectClassification failed, id: %+v, err: %s, rid: %s", id, err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
	}

	err = s.Core.ClassificationOperation().UpdateClassification(ctx.Kit, data, id, cond)
	if err != nil {
		ctx.RespAutoError(err)
		return
	}

	//get AuditLog CurData
	err = objClsAuditLog.WithCurrent(ctx.Kit, id)
	if err != nil {
		blog.Errorf("[api-cls] find Current objectClassification failed, id: %+v, err: %s, rid: %s", id, err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
	}

	//package audit response
	err = objClsAuditLog.SaveAuditLog(ctx.Kit, metadata.AuditUpdate)
	if err != nil {
		ctx.RespAutoError(err)
	}

	ctx.RespEntity(nil)
}

// DeleteClassification delete the object classification
func (s *Service) DeleteClassification(ctx *rest.Contexts) {
	cond := condition.CreateCondition()
	id, err := strconv.ParseInt(ctx.Request.PathParameter("id"), 10, 64)
	if nil != err {
		blog.Errorf("[api-cls] failed to parse the path params id(%s), error info is %s , rid: %s", ctx.Request.PathParameter("id"), err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	objClsAuditLog := operation.NewObjectClsAudit(s.Engine.CoreAPI)
	//get AuditLog PreData
	err = objClsAuditLog.WithPrevious(ctx.Kit, id)
	if err != nil {
		blog.Errorf("[api-cls] find Previous object group failed, id: %+v, err: %s, rid: %s", id, err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
	}

	md := new(MetaShell)
	if err := ctx.DecodeInto(md); err != nil {
		ctx.RespAutoError(err)
		return
	}
	err = s.Core.ClassificationOperation().DeleteClassification(ctx.Kit, id, cond, md.Metadata)
	if nil != err {
		blog.Errorf("[api-cls] failed to parse the path params id(%s), error info is %s , rid: %s", ctx.Request.PathParameter("id"), err.Error(), ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	//package audit response
	err = objClsAuditLog.SaveAuditLog(ctx.Kit, metadata.AuditDelete)
	if err != nil {
		ctx.RespAutoError(err)
	}
	ctx.RespEntity(nil)
}
