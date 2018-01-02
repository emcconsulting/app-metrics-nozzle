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

package service

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	// As original source of restgate doesn't resolve "gopkg.in/unrolled/render.v1" dependency, the resource is maintained locally.
	"app-metrics-nozzle/restgate"
	
	"github.com/gorilla/context"
	"net/http"
)

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {

	formatter := render.New(render.Options{
		IndentJSON: true,
	})

	n := negroni.Classic()
	mx := mux.NewRouter()

	initRoutes(mx, formatter)

	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	//Create subrouters
	secureRouter := mux.NewRouter()
	secureRouter.HandleFunc("/api/apps/{org}/{space}/{app}", appHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/apps/{org}/{space}", appSpaceHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/apps/{org}", appOrgHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/apps", appAllHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/orgs/{org}", orgDetailsHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/orgs", orgsHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/spaces/{space}", spaceDetailsHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/spaces", spaceHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/report/email", generateAllReportHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/report/email/{org}", generateOrgReportHandler(formatter)).Methods("GET")
	secureRouter.HandleFunc("/api/report/email/{org}/{space}", generateSpaceReportHandler(formatter)).Methods("GET")
	
	//Secure the endpoints
	negRest := negroni.New()
	negRest.Use(restgate.New("X-Auth-Key", "X-Auth-Secret", restgate.Static, restgate.Config{Context: C, Key: []string{"12345"}, Secret: []string{"secret"}, HTTPSProtectionOff: true}))
	negRest.UseHandler(secureRouter)

	// Add subrouter to main route
	// These endpoints are protected by RestGate via hardcoded KEYs
	mx.Handle("/api/apps/{org}/{space}/{app}", negRest)
	mx.Handle("/api/apps/{org}/{space}", negRest)
	mx.Handle("/api/apps/{org}", negRest)
	mx.Handle("/api/apps", negRest)
	mx.Handle("/api/orgs/{org}", negRest)
	mx.Handle("/api/orgs", negRest)
	mx.Handle("/api/spaces/{space}", negRest)
	mx.Handle("/api/spaces", negRest)
	mx.Handle("/api/report/email", negRest)
	mx.Handle("/api/report/email/{org}", negRest)
	mx.Handle("/api/report/email/{org}/{space}", negRest)
}

//Optional Context - If not required, remove 'Context: C' or alternatively pass nil (see above)
//NB: Endpoint handler can determine the key used to authenticate via: context.Get(r, 0).(string)
func C(r *http.Request, authenticatedKey string) {
	context.Set(r, 0, authenticatedKey) // Read http://www.gorillatoolkit.org/pkg/context about setting arbitary context key
}
