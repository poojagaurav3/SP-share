package app

import (
	"html"

	"github.com/revel/revel"
	"github.com/sp-share/app/auth"
	"github.com/sp-share/app/common"
	"github.com/sp-share/app/controllers"
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.BeforeAfterFilter,       // Call the before and after filter functions
		revel.ActionInvoker,           // Invoke the action.

	}

	// Auth Interceptor
	revel.InterceptFunc(auth.Authenticate, revel.BEFORE, &controllers.Home{})
	revel.InterceptFunc(auth.Authenticate, revel.BEFORE, &controllers.Group{})
	revel.InterceptFunc(auth.Authenticate, revel.BEFORE, &controllers.Requests{})
	revel.InterceptFunc(auth.Authenticate, revel.BEFORE, &controllers.Item{})
	revel.InterceptFunc(auth.Authenticate, revel.BEFORE, &controllers.Limit{})

	revel.TemplateFuncs["increment"] = func(a int) int {
		return a + 1
	}

	revel.TemplateFuncs["wfstr"] = func(a int) string {
		if a < 0 || a > 2 {
			return "-"
		}
		wf := common.WorkflowStatus(a)
		return wf.GetString()
	}

	revel.TemplateFuncs["isimg"] = func(itemType int) bool {
		if itemType < 0 || itemType > 2 {
			return false
		}
		if common.ItemType(itemType) == common.ItemTypePictures {
			return true
		}

		return false
	}

	revel.TemplateFuncs["isvideo"] = func(itemType int) bool {
		if itemType < 0 || itemType > 2 {
			return false
		}
		if common.ItemType(itemType) == common.ItemTypeVideos {
			return true
		}

		return false
	}

	revel.TemplateFuncs["unescape"] = func(str string) string {
		return html.UnescapeString(str)
	}

	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)
}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")
	c.Response.Out.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}
