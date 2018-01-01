package usageevents

import (
	"fmt"
	"app-metrics-nozzle/domain"
	"app-metrics-nozzle/api"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
)

func ReloadApps(cachedApps []caching.App) {
	logger.Println("Start filling app/space/org cache.")
	for idx := range cachedApps {

		org := cachedApps[idx].OrgName
		space := cachedApps[idx].SpaceName
		app := cachedApps[idx].Name
		key := GetMapKeyFromAppData(org, space, app)

		appId := cachedApps[idx].Guid
		name := cachedApps[idx].Name

		appDetail := domain.App{GUID:appId, Name:name}
		api.AnnotateWithCloudControllerData(&appDetail)
		
		appDetail.LastEventTime = AppDetails[key].LastEventTime
		
		AppDetails[key] = appDetail
		logger.Println(fmt.Sprintf("Registered [%s]", key))
	}

	logger.Println(fmt.Sprintf("Done filling cache! Found [%d] Apps", len(cachedApps)))
}
