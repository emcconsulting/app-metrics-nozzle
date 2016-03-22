package usageevents

import (
	"fmt"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"

	"sync"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

// Event is a struct represented an event augmented/decorated with corresponding app/space/org data.
type Event struct {
	//Fields         logrus.Fields `json:"fields"`
	Msg            string `json:"message"`
	Type           string `json:"event_type"`
	Origin         string `json:"origin"`
	AppID          string `json:"app_id"`
	Timestamp      int64  `json:"timestamp"`
	SourceType     string `json:"source_type"`
	MessageType    string `json:"message_type"`
	SourceInstance string `json:"source_instance"`
	AppName        string `json:"app_name"`
	OrgName        string `json:"org_name"`
	SpaceName      string `json:"space_name"`
	OrgID          string `json:"org_id"`
	SpaceID        string `json:"space_id"`
}

/*
"origin":          msg.GetOrigin(),
"cf_app_id":       logMessage.GetAppId(),
"timestamp":       logMessage.GetTimestamp(),
"source_type":     logMessage.GetSourceType(),
"message_type":    logMessage.GetMessageType().String(),
"source_instance": logMessage.GetSourceInstance(),*/

// ApplicationStat represents the observed metadata about an app, e.g. last router event time, etc.
type ApplicationStat struct {
	LastEventTime int64  `json:"last_event_time"`
	LastEvent     Event  `json:"last_event"`
	EventCount    int64  `json:"event_count"`
	AppName       string `json:"app_name"`
	OrgName       string `json:"org_name"`
	SpaceName     string `json:"space_name"`
}

// ApplicationDetail represents a time snapshot of the RPS and elapsed time since last event for an app
type ApplicationDetail struct {
	Stats                 ApplicationStat `json:"stats"`
	RequestsPerSecond     float64         `json:"req_per_second"`
	ElapsedSinceLastEvent int64           `json:"elapsed_since_last_event"`
}

var mutex sync.Mutex

// AppStats is a map of app names to collected stats.
var AppStats = make(map[string]ApplicationStat)

var feedStarted int64

// ProcessEvents churns through the firehose channel, processing incoming events.
func ProcessEvents(in chan *events.Envelope) {
	feedStarted = time.Now().UnixNano()
	for msg := range in {
		processEvent(msg)
	}
}

func processEvent(msg *events.Envelope) {
	eventType := msg.GetEventType()

	var event Event
	if eventType == events.Envelope_LogMessage {
		event = LogMessage(msg)
		if event.SourceType == "RTR" {
			event.AnnotateWithAppData()
			updateAppStat(event)
		}
	}
	//fmt.Println("tick")
}

// CalculateDetailedStat takes application stats, uses the clock time, and calculates elapsed times and requests/second.
func CalculateDetailedStat(stat ApplicationStat) (detail ApplicationDetail) {
	detail.Stats = stat
	if len(stat.LastEvent.Type) > 0 {
		eventElapsed := time.Now().UnixNano() - stat.LastEventTime
		detail.ElapsedSinceLastEvent = eventElapsed / 1000000000
		totalElapsed := time.Now().UnixNano() - feedStarted
		elapsedSeconds := totalElapsed / 1000000000
		detail.RequestsPerSecond = float64(stat.EventCount) / float64(elapsedSeconds)
	}
	return
}

// GetMapKeyFromAppData converts the combo of an app, space, and org into a hashmap key
func GetMapKeyFromAppData(orgName string, spaceName string, appName string) string {
	return fmt.Sprintf("%s/%s/%s", orgName, spaceName, appName)
}

func updateAppStat(logEvent Event) {
	appName := logEvent.AppName
	appOrg := logEvent.OrgName
	appSpace := logEvent.SpaceName

	appKey := GetMapKeyFromAppData(appOrg, appSpace, appName)
	appStat := AppStats[appKey]
	appStat.LastEventTime = time.Now().UnixNano()
	appStat.EventCount++
	appStat.AppName = appName
	appStat.SpaceName = appSpace
	appStat.OrgName = appOrg
	appStat.LastEvent = logEvent
	AppStats[appKey] = appStat
}

func getAppInfo(appGUID string) caching.App {
	if app := caching.GetAppInfo(appGUID); app.Name != "" {
		return app
	}
	caching.GetAppByGuid(appGUID)

	return caching.GetAppInfo(appGUID)
}

// LogMessage augments a raw message Envelope with log message metadata.
func LogMessage(msg *events.Envelope) Event {
	logMessage := msg.GetLogMessage()

	return Event{
		Origin:         msg.GetOrigin(),
		AppID:          logMessage.GetAppId(),
		Timestamp:      logMessage.GetTimestamp(),
		SourceType:     logMessage.GetSourceType(),
		SourceInstance: logMessage.GetSourceInstance(),
		MessageType:    logMessage.GetMessageType().String(),
		Msg:            string(logMessage.GetMessage()),
		Type:           msg.GetEventType().String(),
	}
}

// AnnotateWithAppData adds application specific details to an event by looking up the GUID in the cache.
func (e *Event) AnnotateWithAppData() {

	cfAppID := e.AppID
	appGUID := ""
	if cfAppID != "" {
		appGUID = fmt.Sprintf("%s", cfAppID)
	}

	if appGUID != "<nil>" && cfAppID != "" {
		appInfo := getAppInfo(appGUID)
		cfAppName := appInfo.Name
		cf_space_id := appInfo.SpaceGuid
		cf_space_name := appInfo.SpaceName
		cf_org_id := appInfo.OrgGuid
		cf_org_name := appInfo.OrgName

		if cfAppName != "" {
			e.AppName = cfAppName
		}

		if cf_space_id != "" {
			e.SpaceID = cf_space_id
		}

		if cf_space_name != "" {
			e.SpaceName = cf_space_name
		}

		if cf_org_id != "" {
			e.OrgID = cf_org_id
		}

		if cf_org_name != "" {
			e.OrgName = cf_org_name
		}
	}
}
