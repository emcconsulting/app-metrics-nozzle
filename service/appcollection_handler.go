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
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"app-metrics-nozzle/usageevents"
	"github.com/unrolled/render"
	"strings"
	"app-metrics-nozzle/domain"
)

func appAllHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		formatter.JSON(w, http.StatusOK, usageevents.AppDetails)
	}
}

func appOrgHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")

		vars := mux.Vars(req)
		org := vars["org"]
		searchKey := fmt.Sprintf("%s/", org)

		searchApps(searchKey, w, formatter)
	}
}

func appSpaceHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")

		vars := mux.Vars(req)
		org := vars["org"]
		space := vars["space"]
		searchKey := fmt.Sprintf("%s/%s/", org, space)

		searchApps(searchKey, w, formatter)
	}
}

func searchApps(searchKey string, w http.ResponseWriter, formatter *render.Render) {
	allAppDetails := usageevents.AppDetails
	foundApps := make(map[string]domain.App)

	for idx, appDetail := range allAppDetails {
		if strings.HasPrefix(idx, searchKey) {
			foundApps[idx] = appDetail
		}
	}

	if 0 < len(foundApps) {
		formatter.JSON(w, http.StatusOK, foundApps)
	} else {
		formatter.JSON(w, http.StatusNotFound, "No such app")
	}
}

//New deep structure with all application details
func appHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")

		vars := mux.Vars(req)
		app := vars["app"]
		org := vars["org"]
		space := vars["space"]
		key := usageevents.GetMapKeyFromAppData(org, space, app)
		stat, exists := usageevents.AppDetails[key]

		if exists {
			//todo calc needed statistics before serving
			formatter.JSON(w, http.StatusOK, stat)
		} else {
			formatter.JSON(w, http.StatusNotFound, "No such app")
		}
	}
}
