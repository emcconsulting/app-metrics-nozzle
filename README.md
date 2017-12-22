# App Metrics Nozzle

This is a nozzle for the Cloud Foundry firehose component. It will ingest router events for every application it can detect and use the timestamps on those events to compute usage metrics. The first usage for this nozzle will be to determine if an application is unused by tracking the last time a request was routed to it.
It will also make REST api calls to Cloud Controller and cache application, and user data for the interval specified in configuration.

## REST API
This application exposes a RESTful API that allows consumers to query application usage statistics. This API is described in the following table:

| Resource        | Method           | Description  |
| --- | --- | --- |
| `/api/apps` | GET | Queries the list of all _deployed_ applications. This is all applications in all organizations, in all spaces. The results will be a map whose key is a string of the format `[org]/[space]/[app name]`. |
| `/api/apps/[org]/[space]/[app]` | GET | Obtains application detail information, including time-based usage statistics as of the time of request, including elapsed time (in seconds) since the last event was received, and the requests per second for the app. |
| `/api/apps/[org]/[space]` | GET | Obtains application details deployed in specified space. |
| `/api/apps/[org]` | GET | Obtains application details deployed in specified organization. |
| `/api/orgs` | GET | Obtains names and guids of all organizations. |
| `/api/orgs/[org]` | GET | Obtains name and guid of an organization. |
| `/api/spaces` | GET | Returns a list of spaces. |
| `/api/spaces/[space]` | GET | Returns space details. |

### JSON Payloads
This is a sample of what the JSON response looks like for the app `/api/apps`:

```javascript

{
	"pcfdev-org/pcfdev-space/app-metrics-nozzle": {
		"guid": "70df4964-60a5-4caa-ba44-99f04f24b778",
		"name": "app-metrics-nozzle",
		"organization": {
			"id": "d826fe71-aeec-4775-8d83-5bf01cebb563",
			"name": "pcfdev-org"
		},
		"event_count": 5,
		"last_event_time": 1513801825226113000,
		"requests_per_second": 0.03067484662576687,
		"elapsed_since_last_event": 0,
		"space": {
			"id": "43bb7404-2ab2-4a56-a426-27d48b1c6958",
			"name": "pcfdev-space"
		},
		"state": "STARTED"
	}
}
,
"org/space/app" : {},
```

If the `last_event_time` field is `0` that indicates that no _router_ events for that application have been discovered _since the nozzle was started_.

## Installation
Run glide install to pull dependencies into vendor directory.
To install this application, it should be run as an app within Cloud Foundry. So, the first thing you'll need to do is push the app. There is a `manifest.yml` already included in the project, so you can just do:

```
cf push app-metrics-nozzle --no-start
```

The `no-start` is important because we have not yet defined the environment variables that allow the application to connect to the Firehose and begin monitoring router requests. We want to end up with a set of environment variables that looks like this when we issue a `cf env app-metrics-nozzle` command:

```
User-Provided:
API_ENDPOINT: https://api.local.pcfdev.io
DOPPLER_ENDPOINT: wss://doppler.local.pcfdev.io:443
CF_PULL_TIME: 9999s
FIREHOSE_PASSWORD: (this is a secret)
FIREHOSE_SUBSCRIPTION_ID: app-metrics-nozzle
FIREHOSE_USER: (this is also secret)
SKIP_SSL_VALIDATION: true
```
Once you've set these environment variables with `cf set-env (app) (var) (value)` you can just start the application usage nozzle via `cf start`. Make sure the application has come up by hitting the API endpoint. Depending on how large of a foundation in which it was deployed, it can take _several minutes_ for the cache of application metadata to fill up.

DOPPLER_ENDPOINT can be obtained by running
```bash
cf curl /v2/info
```

