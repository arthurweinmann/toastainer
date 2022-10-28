import "local://includes/config.js";
import "local://includes/routes/user.js";

const routes = [
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/"
        
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/create-toaster"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/edit-toaster"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/toaster"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/my-toasters"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/custom-domains"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/subdomains"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/create-domain"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/edit-domain"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/create-subdomain"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/pricing"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/account"
    },
    {
        authenticated: true,
        userAccess: ["all"],
        path: "/settings"
    },
    {
        authenticated: false,
        userAccess: ["all"],
        path: "/signup"
    },
    {
        authenticated: false,
        userAccess: ["all"],
        path: "/login"
    },
    {
        authenticated: false,
        userAccess: ["all"],
        path: "/forget-password"
    },
    {
        authenticated: false,
        userAccess: ["all"],
        path: "/reset-password"
    },
    {
        authenticated: false,
        userAccess: ["all"],
        path: "/terms-of-service"
    },
    {
        authenticated: false,
        userAccess: ["all"],
        path: "/privacy"
    },
];

/**
 * 
 *  ROUTER IS IN DEVELOPMENT, NEED TO BE ENHANCE
 */
class Router {
    constructor() {
        this.routes = routes;
        this.loadInitialRoute();
    }

    loadInitialRoute() {
        // path segments for the route which should load initially.
        let lastChar = window.location.pathname.slice(-1);

        // handle if we have a route ended by a "/", excepted root - example: /dashboard/ => /dashboard
        let path = lastChar === "/" && window.location.pathname !== "/" ? window.location.pathname.slice(0, -1) : window.location.pathname;

        const pathnameSplit = path.split('/');
        const pathSegments = pathnameSplit.length > 1 ? pathnameSplit.slice(1) : '';

        this.loadRoute(...pathSegments);
    }

    loadRoute(...urlSegments) {
        const matchedRoute = this.matchUrlToRoute(urlSegments);
        console.log("Router matchedRoute", matchedRoute);
        console.log("Router user", USER);

        if (matchedRoute) {
            if (USER) {     // logged in
                let userType = getUserType(USER);

                if (!matchedRoute.authenticated && !matchedRoute.exception) {      // if user acces to unauthenticated route, redirect
                    window.location = "/";
                }

                if (matchedRoute.userAccess.includes(userType) || matchedRoute.exception || matchedRoute.userAccess.includes("all")) {           // get access
                    // We pass an empty object and an empty string as the historyState and title arguments, but their values do not really matter here.
                    // const url = `/${urlSegments.join('/')}`;
                    // history.pushState({}, '', url);

                    // append the given template to the DOM inside the router outlet.
                    // const routerOutletElement = document.querySelectorAll('[data-router-outlet]')[0];
                    // routerOutletElement.innerHTML = matchedRoute.getTemplate(matchedRoute.params);
                }
                else {    // no access, redirect user
                    window.location = "/";
                }
            }
            else {          // not logged in
                if (matchedRoute.authenticated) {
                    window.location = "/login";
                }
            }
        }
        else {
            // no page found
        }
    }

    matchUrlToRoute(urlSegments) {
        const routeParams = {};
        const matchedRoute = this.routes.find(route => {
            // assume that the route path always starts with a slash, and so 
            // the first item in the segments array  will always be an empty
            // string. Slice the array at index 1 to ignore this empty string.
            const routePathSegments = route.path.split('/').slice(1);

            if (routePathSegments.length !== urlSegments.length) {          // If different numbers of segments, then the route does not match the URL.
                return false;
            }

            // If each segment in the url matches the corresponding segment in the route path, 
            // or the route path segment starts with a ':' then the route is matched.
            const match = routePathSegments.every((routePathSegment, i) => {
                return routePathSegment === urlSegments[i] || routePathSegment[0] === ':';
            });

            // if route matches the URL, pull out any params from the URL.
            if (match) {
                routePathSegments.forEach((segment, i) => {
                    if (segment[0] === ':') {
                        const propName = segment.slice(1);
                        routeParams[propName] = decodeURIComponent(urlSegments[i]);
                    }
                });
            }
            return match;
        });

        return { ...matchedRoute, params: routeParams };
    }

    sanitizeURL(url) {

    }
}