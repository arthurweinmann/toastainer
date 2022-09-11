package dynamicroutes

import (
	"net/http"
	"strings"
)

/*
	- Set if AWS SES used in email/awsses package: sys/rejectemail
*/
var unauthedDynamicRoutes = map[string]interface{}{}

func AddUnauthedDynamicRoute(pathWithoutLeadingSlash string, handler func(w http.ResponseWriter, r *http.Request)) {
	spl := strings.Split(pathWithoutLeadingSlash, "/")
	tmp := unauthedDynamicRoutes
	for i := 0; i < len(spl); i++ {
		_, ok := tmp[spl[i]]
		if i < len(spl)-1 {
			if !ok {
				tmp[spl[i]] = map[string]interface{}{}
			} else {
				_, ok = tmp[spl[i]].(map[string]interface{})
				if !ok {
					panic("handlers in dynamic routes are only supported for the last part of the route")
				}
			}
			tmp = tmp[spl[i]].(map[string]interface{})
		} else {
			if ok {
				panic("a handler is already registered or another dynamic route is using the same route start path")
			}
			tmp[spl[i]] = handler
		}
	}
}

// GetDynamicRoute returns nil if it did not find the route
func GetDynamicRoute(spath []string) func(w http.ResponseWriter, r *http.Request) {
	tmp := unauthedDynamicRoutes
	for i := 0; i < len(spath); i++ {
		next, ok := tmp[spath[i]]
		if !ok {
			return nil
		}
		if i < len(spath)-1 {
			switch t := next.(type) {
			case map[string]interface{}:
				tmp = t
			default:
				return nil
			}
		} else {
			switch t := next.(type) {
			case func(w http.ResponseWriter, r *http.Request):
				return t
			default:
				return nil
			}
		}
	}

	return nil
}
