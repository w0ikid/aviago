package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type WhoamiResponse struct {
	Identity struct {
		ID     string `json:"id"`
		Traits struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"traits"`
	} `json:"identity"`
}

// middleware, который проверяет, есть ли валидная сессия у пользователя
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// достаем cookie из запроса
		cookie, err := r.Cookie("ory_kratos_session") // имя сессии Kratos по дефолту
		if err != nil {
			http.Error(w, "Unauthorized (no session cookie)", http.StatusUnauthorized)
			return
		}

		// пробрасываем cookie в запрос к Kratos
		req, _ := http.NewRequest("GET", "http://127.0.0.1:4433/sessions/whoami", nil)
		req.AddCookie(cookie)

		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			http.Error(w, "Unauthorized (invalid session)", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var whoami WhoamiResponse
		if err := json.Unmarshal(body, &whoami); err != nil {
			http.Error(w, "Unauthorized (bad whoami response)", http.StatusUnauthorized)
			return
		}

		// ✅ если дошли сюда, значит пользователь авторизован
		fmt.Printf("Authorized user: %s (%s)\n", whoami.Identity.Traits.Name, whoami.Identity.Traits.Email)

		// передаем дальше в хендлер
		next.ServeHTTP(w, r)
	})
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Welcome!</h1><p>You are on /welcome page 🚀</p>")
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/welcome", authMiddleware(http.HandlerFunc(welcomeHandler)))

	fmt.Println("Server is running on http://127.0.0.1:4455/welcome")
	if err := http.ListenAndServe(":4455", mux); err != nil {
		panic(err)
	}
}
