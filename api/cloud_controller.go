/*
Copyright 2016 Pivotal

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"os"
	"log"
	"app-metrics-nozzle/domain"
)

var logger = log.New(os.Stdout, "", 0)
var Client CFClientCaller

type CFClientCaller interface {
	AppByGuid(guid string) (cfclient.App, error)
	ListSpaces() ([]cfclient.Space, error)
	ListOrgs() ([]cfclient.Org, error)
	ListApps() ([]cfclient.App, error)
	AppSpace(app cfclient.App) (cfclient.Space, error)
	SpaceOrg(space cfclient.Space) (cfclient.Org, error)
}

func AppByGuidVerify(guid string) (cfclient.App) {
	app, _ := Client.AppByGuid(guid)
	return app
}

func AnnotateWithCloudControllerData(app *domain.App) {

	ccAppDetails, _ := Client.AppByGuid(app.GUID)

	space, _ := Client.AppSpace(ccAppDetails)
	org, _ := Client.SpaceOrg(space)

	app.Organization.ID = org.Guid
	app.Organization.Name = org.Name

	app.Space.ID = space.Guid
	app.Space.Name = space.Name

	app.State = ccAppDetails.State
}

func SpacesDetailsFromCloudController() (Spaces []cfclient.Space) {
	spaces, _ := Client.ListSpaces()
	return spaces
}

func OrgsDetailsFromCloudController() (Orgs []cfclient.Org) {
	orgs, _ := Client.ListOrgs()
	return orgs
}




