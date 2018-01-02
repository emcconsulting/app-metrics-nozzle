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
	Space                 struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"space"`
	State                 string `json:"state"`
}


