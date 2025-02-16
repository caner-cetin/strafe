package endpoints

import "net/http"

func Health(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=NuXjeEC2XOA", http.StatusFound)
}
