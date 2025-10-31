package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/helpers"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/middleware"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Send whatsapp API over http",
	Long:  `This application is from clone https://github.com/aldinokemal/go-whatsapp-web-multidevice`,
	Run:   restServer,
}

func init() {
	rootCmd.AddCommand(restCmd)
}

// setupNextJSServe sets up Next.js to be served via /web route
func setupNextJSServe(app *fiber.App) {
	nextjsPath := filepath.Join(".", "whatsapp-web")

	// Check if whatsapp-web directory exists
	if _, err := os.Stat(nextjsPath); os.IsNotExist(err) {
		logrus.Warn("whatsapp-web directory not found, skipping Next.js setup")
		return
	}

	// Always start Next.js development server for simplicity
	go startNextJSDev(nextjsPath)

	// Serve Next.js app via /web route using reverse proxy
	app.All("/web/*", func(c *fiber.Ctx) error {
		// Handle API routes in Next.js - proxy them to backend Go API
		if strings.HasPrefix(c.Path(), "/web/api/") {
			return proxyToNextJSAPI(c)
		}

		// Proxy all other /web/* requests to Next.js development server
		return proxyToNextJS(c)
	})

	// Proxy Next.js static assets that are referenced without the /web prefix
	app.All("/_next/*", func(c *fiber.Ctx) error {
		return proxyToNextJS(c)
	})

	app.All("/:file.svg", func(c *fiber.Ctx) error {
		return proxyToNextJS(c)
	})
	app.All("/:file.png", func(c *fiber.Ctx) error {
		return proxyToNextJS(c)
	})
	app.All("/:file.jpg", func(c *fiber.Ctx) error {
		return proxyToNextJS(c)
	})

	// Proxy Next.js HMR and dev helpers
	app.All("/__next*/*", func(c *fiber.Ctx) error {
		return proxyToNextJS(c)
	})

	logrus.Info("Next.js app configured to serve via /web route")
}

// startNextJSDev starts Next.js in development mode as fallback
func startNextJSDev(nextjsPath string) {
	logrus.Info("Starting Next.js development server...")

	// Check if node_modules exists
	nodeModulesPath := filepath.Join(nextjsPath, "node_modules")
	if _, err := os.Stat(nodeModulesPath); os.IsNotExist(err) {
		logrus.Info("Installing Next.js dependencies...")
		installCmd := exec.Command("npm", "install")
		installCmd.Dir = nextjsPath
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			logrus.Error("Failed to install Next.js dependencies: ", err.Error())
			return
		}
	}

	// Prepare environment variables for Next.js
	env := os.Environ()

	// Export APP_BASIC_AUTH to Next.js (shared with backend)
	if basicAuth := os.Getenv("APP_BASIC_AUTH"); basicAuth != "" {
		env = append(env, fmt.Sprintf("APP_BASIC_AUTH=%s", basicAuth))
		env = append(env, fmt.Sprintf("NEXT_PUBLIC_APP_BASIC_AUTH=%s", basicAuth))
	}

	// Export backend URL to Next.js
	env = append(env, fmt.Sprintf("NEXT_PUBLIC_WHATSAPP_API_URL=http://localhost:%s", config.AppPort))

	// Start Next.js development server with environment variables
	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = nextjsPath
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logrus.Error("Failed to start Next.js: ", err.Error())
		return
	}

	logrus.Info("Next.js development server started - access via http://localhost:3001")
}

// proxyToNextJSAPI proxies Next.js API requests to local backend
func proxyToNextJSAPI(c *fiber.Ctx) error {
	// Extract the API path without /web prefix
	apiPath := strings.TrimPrefix(c.Path(), "/web")

	// Redirect to the actual backend API endpoints
	// This creates a seamless integration where Next.js API routes work through the Go backend
	targetURL := "http://localhost:" + config.AppPort + apiPath

	// Create a new request to forward
	proxyReq, err := http.NewRequest(c.Method(), targetURL, strings.NewReader(string(c.Body())))
	if err != nil {
		return c.Status(500).SendString("Failed to create proxy request")
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		proxyReq.Header.Set(string(key), string(value))
	})

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return c.Status(502).SendString("Backend server unavailable")
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Set status code
	c.Status(resp.StatusCode)

	// Stream response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	return c.Send(bodyBytes)
}

// proxyToNextJS forwards requests to Next.js development server
func proxyToNextJS(c *fiber.Ctx) error {
	// Strip /web prefix so Next.js receives paths relative to its root
	path := strings.TrimPrefix(c.Path(), "/web")
	if path == "" {
		path = "/"
	}
	targetURL := "http://localhost:3001" + path

	// Create a new request to forward
	proxyReq, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(500).SendString("Failed to create proxy request")
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		proxyReq.Header.Set(string(key), string(value))
	})

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return c.Status(502).SendString("Next.js server unavailable")
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Set status code
	c.Status(resp.StatusCode)

	// Stream response body
	bodyBytes, _ := io.ReadAll(resp.Body)
	return c.Send(bodyBytes)
}

func restServer(_ *cobra.Command, _ []string) {
	engine := html.NewFileSystem(http.FS(EmbedIndex), ".html")
	engine.AddFunc("isEnableBasicAuth", func(token any) bool {
		return token != nil
	})
	app := fiber.New(fiber.Config{
		Views:                   engine,
		EnableTrustedProxyCheck: true,
		BodyLimit:               int(config.WhatsappSettingMaxVideoSize),
		Network:                 "tcp",
	})

	app.Static(config.AppBasePath+"/statics", "./statics")
	app.Use(config.AppBasePath+"/components", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/components",
		Browse:     true,
	}))
	app.Use(config.AppBasePath+"/assets", filesystem.New(filesystem.Config{
		Root:       http.FS(EmbedViews),
		PathPrefix: "views/assets",
		Browse:     true,
	}))

	app.Use(middleware.Recovery())
	if config.AppDebug {
		app.Use(logger.New())
	}

	// CORS middleware must be before auth
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000,http://127.0.0.1:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Apply Basic Auth if configured
	if len(config.AppBasicAuthCredential) > 0 {
		account := make(map[string]string)
		for _, basicAuth := range config.AppBasicAuthCredential {
			ba := strings.Split(basicAuth, ":")
			if len(ba) != 2 {
				logrus.Fatalln("Basic auth is not valid, please this following format <user>:<secret>")
			}
			account[ba[0]] = ba[1]
		}

		// Use Fiber's built-in Basic Auth middleware
		app.Use(basicauth.New(basicauth.Config{
			Users: account,
		}))
	}

	// Custom middleware to capture auth token for API usage
	app.Use(middleware.BasicAuth())

	// Setup Next.js to be served via /web route
	setupNextJSServe(app)

	// Create base path group or use app directly
	var apiGroup fiber.Router = app
	if config.AppBasePath != "" {
		apiGroup = app.Group(config.AppBasePath)
	}

	// Rest
	rest.InitRestApp(apiGroup, appUsecase)
	rest.InitRestChat(apiGroup, chatUsecase)
	rest.InitRestSend(apiGroup, sendUsecase)
	rest.InitRestUser(apiGroup, userUsecase)
	rest.InitRestMessage(apiGroup, messageUsecase)
	rest.InitRestGroup(apiGroup, groupUsecase)
	rest.InitRestNewsletter(apiGroup, newsletterUsecase)

	apiGroup.Get("/", func(c *fiber.Ctx) error {
		return c.Render("views/index", fiber.Map{
			"AppHost":        fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname()),
			"AppVersion":     config.AppVersion,
			"AppBasePath":    config.AppBasePath,
			"BasicAuthToken": c.UserContext().Value(middleware.AuthorizationValue("BASIC_AUTH")),
			"MaxFileSize":    humanize.Bytes(uint64(config.WhatsappSettingMaxFileSize)),
			"MaxVideoSize":   humanize.Bytes(uint64(config.WhatsappSettingMaxVideoSize)),
			"WebURL":         fmt.Sprintf("%s://%s%s/web", c.Protocol(), c.Hostname(), config.AppBasePath),
		})
	})

	websocket.RegisterRoutes(apiGroup, appUsecase)
	go websocket.RunHub()

	// Set auto reconnect to whatsapp server after booting
	go helpers.SetAutoConnectAfterBooting(appUsecase)
	// Set auto reconnect checking
	go helpers.SetAutoReconnectChecking(whatsappCli)

	logrus.Info("🚀 WhatsApp API Server starting...")
	logrus.Info("📱 Web Interface: http://localhost:" + config.AppPort + "/web")
	logrus.Info("🔗 API Documentation: http://localhost:" + config.AppPort)

	if err := app.Listen(":" + config.AppPort); err != nil {
		logrus.Fatalln("Failed to start: ", err.Error())
	}
}
