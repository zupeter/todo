package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

func (s *Server) RegisterRoutes() http.Handler {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	router.GET("/", s.HelloWorldHandler)

	router.GET("/health", s.healthHandler)

	router.GET("/auth/{provider}/callback", s.getAuthCallbackHandler)

	router.GET("/logout/{provider}", s.getAuthLogoutHandler)

	router.GET("/auth/{provider}")

	return router
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) getAuthCallbackHandler(c *gin.Context) {
	var res http.ResponseWriter = c.Writer
	var r *http.Request = c.Request

	provider := chi.URLParam(r, "provider")
	req := r.WithContext(context.WithValue(context.Background(), "provider", provider))

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Fprintln(res, err)
		return
	}
	fmt.Println(user)

	http.Redirect(res, r, "http://localhost:5173", http.StatusFound)
}

func (s *Server) getAuthLogoutHandler(c *gin.Context) {
	var res http.ResponseWriter = c.Writer
	var r *http.Request = c.Request

	provider := chi.URLParam(r, "provider")
	req := r.WithContext(context.WithValue(context.Background(), "provider", provider))

	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) getAuthHandler(c *gin.Context) {
	var res http.ResponseWriter = c.Writer
	var r *http.Request = c.Request

	provider := chi.URLParam(r, "provider")
	req := r.WithContext(context.WithValue(context.Background(), "provider", provider))

	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(res, gothUser)
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
