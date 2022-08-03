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

package app

import (
	"context"
	"fmt"
	"time"

	"configcenter/src/ac/extensions"
	"configcenter/src/ac/iam"
	"configcenter/src/common"
	"configcenter/src/common/auth"
	"configcenter/src/common/backbone"
	"configcenter/src/common/blog"
	"configcenter/src/common/types"
	"configcenter/src/scene_server/proc_server/app/options"
	"configcenter/src/scene_server/proc_server/logics"
	"configcenter/src/scene_server/proc_server/service"
)

// Run TODO
func Run(ctx context.Context, cancel context.CancelFunc, op *options.ServerOption) error {

	svrInfo, err := types.NewServerInfo(op.ServConf)
	if err != nil {
		blog.Errorf("fail to new server information. err: %s", err.Error())
		return fmt.Errorf("make server information failed, err:%v", err)
	}

	procSvr := new(service.ProcServer)

	input := &backbone.BackboneParameter{
		ConfigUpdate: procSvr.OnProcessConfigUpdate,
		ConfigPath:   op.ServConf.ExConfig,
		Regdiscv:     op.ServConf.RegDiscover,
		SrvInfo:      svrInfo,
	}
	engine, err := backbone.NewBackbone(ctx, input)
	if err != nil {
		return fmt.Errorf("new backbone failed, err: %v", err)
	}
	configReady := false
	for sleepCnt := 0; sleepCnt < common.APPConfigWaitTime; sleepCnt++ {
		if nil == procSvr.Config {
			time.Sleep(time.Second)
		} else {
			configReady = true
			break
		}
	}
	if !configReady {
		return fmt.Errorf("configuration item not found")
	}
	mongo, err := engine.WithMongo()
	if err != nil {
		return err
	}
	procSvr.Config.Mongo = &mongo

	procSvr.Config.Auth, err = iam.ParseConfigFromKV("authServer", nil)
	if err != nil {
		blog.Warnf("parse auth center config failed: %v", err)
	}

	iamCli := new(iam.IAM)
	if auth.EnableAuthorize() {
		blog.Info("enable auth center access")
		iamCli, err = iam.NewIAM(procSvr.Config.Auth, engine.Metric().Registry())
		if err != nil {
			return fmt.Errorf("new iam client failed: %v", err)
		}
	} else {
		blog.Infof("disable auth center access")
	}

	procSvr.AuthManager = extensions.NewAuthManager(engine.CoreAPI, iamCli)
	procSvr.Engine = engine
	procSvr.Logic = &logics.Logic{
		Engine: procSvr.Engine,
	}

	err = backbone.StartServer(ctx, cancel, engine, procSvr.WebService(), true)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		blog.Infof("process will exit!")
	}

	return nil
}
