type apiError struct {
	Err    string `json:"error"`
	Status int    `json:"status"`
}

func (err apiError) Error() string {
	return fmt.Sprintf("%s, status: %d", err.Err, err.Status)
}

type apiHandler func(http.ResponseWriter, *http.Request) error

func makeHttpHandler(f apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(apiError); ok {
				writeJson(w, e.Status, e)
			} else {
				writeJson(w, http.StatusInternalServerError, apiError{Err: "Internal server error", Status: http.StatusInternalServerError})
			}
		}
	}
}

func writeJson(w http.ResponseWriter, status int, value any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(value)
}
