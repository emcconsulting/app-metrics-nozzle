package usageevents

import (
	"fmt"
	"app-metrics-nozzle/domain"
	"app-metrics-nozzle/api"
	"app-metrics-nozzle/redis"
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
		
		if AppDetails[key].LastEventTime > 0 {
			appDetail.LastEventTime = AppDetails[key].LastEventTime
			// Update redis cache with new LastEventTime
			redis.Set(key, appDetail.LastEventTime)
		} else{	// Fetch LastEventTime from Redis cache
			appDetail.LastEventTime = redis.Get(key)
		}
		
		AppDetails[key] = appDetail
		logger.Println(fmt.Sprintf("Registered [%s]", key))
	}

	logger.Println(fmt.Sprintf("Done filling cache! Found [%d] Apps", len(cachedApps)))
}
