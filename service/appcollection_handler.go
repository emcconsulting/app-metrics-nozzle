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
	
	"log"
	"os"
	"time"
	"net/smtp"
	"net/mail"
	"encoding/csv"
	"bytes"
	// TODO: this import needs to point to github. fix using glide.yaml file
	"app-metrics-nozzle/email"
	//github.com/scorredoira/email
	"gopkg.in/alecthomas/kingpin.v2"
)

var(
	emailSubject = kingpin.Flag("email-subject", "Report email's subject.").Default("Report").OverrideDefaultFromEnvar("EMAIL_SUBJECT").String()
	emailBody = kingpin.Flag("email-body", "Report email's body.").Default("Please find attachment for the report.").OverrideDefaultFromEnvar("EMAIL_BODY").String()
	emailSender = kingpin.Flag("email-sender", "Report sender name.").Default("Admin").OverrideDefaultFromEnvar("EMAIL_SENDER").String()
	emailReceiver = kingpin.Flag("email-receiver", "Report receiver email address.").Default("user@email.com").OverrideDefaultFromEnvar("EMAIL_RECEIVER").String()
	emailServerHost = kingpin.Flag("email-server-host", "SMTP server address.").Default("localhost").OverrideDefaultFromEnvar("EMAIL_SERVER_HOST").String()
	emailServerPort = kingpin.Flag("email-server-port", "SMTP server port.").Default("25").OverrideDefaultFromEnvar("EMAIL_SERVER_PORT").String()
	emailAttachmentName = kingpin.Flag("email-attachment-name", "Report name.").Default("Report.csv").OverrideDefaultFromEnvar("EMAIL_ATTACHMENT_NAME").String()
	emailUserName = kingpin.Flag("email-user-name", "Report sender's email address.").Default("admin@email.com").OverrideDefaultFromEnvar("EMAIL_USER_NAME").String()
	emailUserPassword = kingpin.Flag("email-user-password", "Report sender's email account password.").Default("password").OverrideDefaultFromEnvar("EMAIL_USER_PASSWORD").String()
	reportTimeZone = kingpin.Flag("report-time-zone", "Time zone of the report").Default("Australia/Sydney").OverrideDefaultFromEnvar("REPORT_TIME_ZONE").String()
	
	timeZoneLocation time.Location
)

var logger = log.New(os.Stdout, "", 0)

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

func generateReportHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		
		reportData := GenerateReport()
		err := SendReport(reportData)
		if err != nil {
			logger.Println(err)
		}
		
		formatter.JSON(w, http.StatusOK, "Report is sent to admin email account. Kindly check.")
	}
}

// Emails the report
func SendReport(reportData []byte) error {
	m := email.NewMessage(*emailSubject, *emailBody)
	m.From = mail.Address{
		Name: *emailSender,
		Address: *emailUserName,
	}
	m.To = []string{*emailReceiver}
		
	m.Attachments[*emailAttachmentName] = &email.Attachment{
		Filename: *emailAttachmentName,
		Data:     reportData,
		Inline:   false,
	}
	
	err := email.Send(*emailServerHost + ":" + *emailServerPort, smtp.PlainAuth("", *emailUserName, *emailUserPassword, *emailServerHost), m)
	if err != nil {
		log.Println(err)
	}
	
	return err;
}

// generateReport returns the report with cache data
func GenerateReport() []byte {
	var rows [][]string
	colhdrs := []string{"Org", "Space", "App Name", "Last accessed time"}
	rows = append(rows, colhdrs)
	
	// get the data from each struct
	for _, v := range usageevents.AppDetails {
		if v.Name == "" {
			continue;
		}
		
        row := make([]string, 0, 4)

		row = append(row, v.Organization.Name)
		row = append(row, v.Space.Name)
		row = append(row, v.Name)
		
		if v.LastEventTime > 0 {
			// time zone for the report generation 
			// TODO: move this timezone declaration to global.
			timeZoneLocation, err := time.LoadLocation(*reportTimeZone)
			if err != nil {
				logger.Println("Error loading timezone. Falling back to server time zone:", err)
		        row = append(row, time.Unix(0, v.LastEventTime).Format("02/01/2006, 15:04:05"))
		    } else{
		    	zoneSpecificTime := time.Unix(0, v.LastEventTime).In(timeZoneLocation)
			    row = append(row, zoneSpecificTime.Format("02/01/2006, 15:04:05"))
		    }
		} else {
			row = append(row, "NEVER")
		}
        rows = append(rows, row)
	}

	var buf1 bytes.Buffer
	w := csv.NewWriter(&buf1)
	w.WriteAll(rows)

	return buf1.Bytes()
}
