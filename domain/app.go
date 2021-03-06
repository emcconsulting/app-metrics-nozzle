package domain

type App struct {
	GUID                  string `json:"guid"`
	Name                  string `json:"name"`
	Organization          struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"organization"`
	EventCount            int64 `json:"event_count"`
	LastEventTime         int64   `json:"last_event_time"`
	RequestsPerSecond     float64      `json:"requests_per_second"`
	ElapsedSinceLastEvent int64    `json:"elapsed_since_last_event"`
	Space                 struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"space"`
	State                 string `json:"state"`
}


