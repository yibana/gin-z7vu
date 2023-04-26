package amazon

type CategoryRank struct {
	ID           string `json:"id"`
	Rank         string `json:"rank"`
	Title        string `json:"title"`
	Url          string `json:"url"`
	Img          string `json:"img"`
	Price        string `json:"price"`
	Rating       string `json:"rating"`
	RatingsCount string `json:"ratingsCount"`
	Path         string `json:"path"`
}
