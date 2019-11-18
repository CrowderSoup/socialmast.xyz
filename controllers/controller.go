package controllers

import (
	"io/ioutil"
	"net/http"

	"github.com/CrowderSoup/socialboat/services"
	"github.com/jinzhu/gorm"
	echo "github.com/labstack/echo/v4"
)

// CustomContextHandler struct with a func for handling the custom context intjection
type CustomContextHandler struct {
	ProfileService *services.ProfileService
}

// NewCustomContextHandler returns a new CustomContextHandler
func NewCustomContextHandler(db *gorm.DB) *CustomContextHandler {
	return &CustomContextHandler{
		ProfileService: services.NewProfileService(db),
	}
}

// Handler injects our custom context
func (h *CustomContextHandler) Handler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		s, err := services.GetSession("Boat", c)
		if err != nil {
			return err
		}

		cc := &BoatContext{
			Context:        c,
			Session:        s,
			ProfileService: h.ProfileService,
		}
		return next(cc)
	}
}

// ManifestHandler handles the manifest
func ManifestHandler(ctx echo.Context) error {
	manifest, err := ioutil.ReadFile("./manifest.webmanifest")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error reading manifest")
	}

	return ctx.Blob(http.StatusOK, "application/manifest+json", manifest)
}

// BoatContext custom echo Context for boat
type BoatContext struct {
	echo.Context

	Session        *services.Session
	ProfileService *services.ProfileService
}

// LoggedIn checks if a contexts session is logged in
func (bc *BoatContext) LoggedIn() bool {
	return bc.Session.LoggedIn()
}

// EnsureLoggedIn ensures a user is logged in, throws error if not
func (bc *BoatContext) EnsureLoggedIn() error {
	if !bc.LoggedIn() {
		return echo.NewHTTPError(http.StatusUnauthorized, "You must be logged in")
	}

	return nil
}

// RedirectIfLoggedIn redirects to given path if logged in
func (bc *BoatContext) RedirectIfLoggedIn(path string) error {
	if bc.LoggedIn() {
		return bc.Redirect(http.StatusSeeOther, path)
	}

	return nil
}

// ReturnView renders a view, adding some data to the return
func (bc *BoatContext) ReturnView(code int, view string, data echo.Map) error {
	// Set "title" if not already set
	if _, ok := data["title"]; !ok {
		data["title"] = "SocialMast"
	}

	// Set "loggedIn" if not already set
	if _, ok := data["loggedIn"]; !ok {
		data["loggedIn"] = bc.LoggedIn()
	}

	// Set "profile" if not already set
	if _, ok := data["profile"]; !ok {
		profile, err := bc.ProfileService.GetFirst()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "couldn't get profile")
		}

		data["profile"] = profile
	}

	return bc.Render(code, view, data)
}