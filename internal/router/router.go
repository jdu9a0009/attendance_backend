package router

import (
	"time"
	"university-backend/foundation/web"
	"university-backend/internal/auth"
	"university-backend/internal/controller/http/v1/file"
	"university-backend/internal/repository/postgres/attendance"
	"university-backend/internal/repository/postgres/department"
	"university-backend/internal/repository/postgres/position"

	"github.com/gin-contrib/cors"
	"github.com/redis/go-redis/v9"

	"university-backend/internal/middleware"
	"university-backend/internal/pkg/repository/postgresql"

	"university-backend/internal/repository/postgres/user"

	attendance_controller "university-backend/internal/controller/http/v1/attendance"
	auth_controller "university-backend/internal/controller/http/v1/auth"
	department_controller "university-backend/internal/controller/http/v1/department"
	position_controller "university-backend/internal/controller/http/v1/position"
	user_controller "university-backend/internal/controller/http/v1/user"
)

type Router struct {
	*web.App
	postgresDB         *postgresql.Database
	redisDB            *redis.Client
	port               string
	auth               *auth.Auth
	fileServerBasePath string
}

func NewRouter(
	app *web.App,
	postgresDB *postgresql.Database,
	redisDB *redis.Client,
	port string,
	auth *auth.Auth,
	fileServerBasePath string,
) *Router {
	return &Router{
		app,
		postgresDB,
		redisDB,
		port,
		auth,
		fileServerBasePath,
	}
}

func (r Router) Init() error {
	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Add other origins if needed
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	// repositories:
	// - postgresql
	userPostgres := user.NewRepository(r.postgresDB)
	departmentPostgres := department.NewRepository(r.postgresDB)
	positionPostgres := position.NewRepository(r.postgresDB)
	attendancePostgres := attendance.NewRepository(r.postgresDB)

	// controller
	userController := user_controller.NewController(userPostgres)
	authController := auth_controller.NewController(userPostgres)
	departmentController := department_controller.NewController(departmentPostgres)
	positionController := position_controller.NewController(positionPostgres)
	attendanceController := attendance_controller.NewController(attendancePostgres)

	fileC := file.NewController(r.App, r.fileServerBasePath)

	// #auth
	r.Post("/api/v1/sign-in", authController.SignIn)
	r.Post("/api/v1/refresh-token", authController.RefreshToken)

	r.GET("/media/*filepath", fileC.File)
	r.HEAD("/media/*filepath", fileC.File)

	// #user
	r.Get("/api/v1/user/list", userController.GetUserList, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/user/:id", userController.GetUserDetailById, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Post("/api/v1/user/create", userController.CreateUser, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Put("/api/v1/user/:id", userController.UpdateUserAll, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Patch("/api/v1/user/:id", userController.UpdateUserColumns, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Delete("/api/v1/user/:id", userController.DeleteUser, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/user/statistics", userController.GetStatistics, middleware.Authenticate(r.auth))
	r.Get("/api/v1/user/monthly", userController.GetMonthlyStatistics, middleware.Authenticate(r.auth))
	r.Get("/api/v1/user/dashboard", userController.GetEmployeeDashboard, middleware.Authenticate(r.auth))

	// #department
	r.Get("/api/v1/department/list", departmentController.GetList, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/department/:id", departmentController.GetDetailById, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Post("/api/v1/department/create", departmentController.Create, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Put("/api/v1/department/:id", departmentController.UpdateAll, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Patch("/api/v1/department/:id", departmentController.UpdateColumns, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Delete("/api/v1/department/:id", departmentController.Delete, middleware.Authenticate(r.auth, auth.RoleAdmin))

	// #position
	r.Get("/api/v1/position/list", positionController.GetList, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/position/:id", positionController.GetDetailById, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Post("/api/v1/position/create", positionController.Create, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Put("/api/v1/position/:id", positionController.UpdateAll, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Patch("/api/v1/position/:id", positionController.UpdateColumns, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Delete("/api/v1/position/:id", positionController.Delete, middleware.Authenticate(r.auth, auth.RoleAdmin))

	// #attendance
	r.Get("/api/v1/attendance/list", attendanceController.GetList, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/attendance/:id", attendanceController.GetDetailById, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Post("/api/v1/attendance/create", attendanceController.Create, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Post("/api/v1/attendance/createbyphone", attendanceController.CreateByPhone, middleware.Authenticate(r.auth))
	r.Patch("/api/v1/attendance/exitbyphone", attendanceController.ExitByPhone, middleware.Authenticate(r.auth))
	r.Put("/api/v1/attendance/:id", attendanceController.UpdateAll, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Patch("/api/v1/attendance/:id", attendanceController.UpdateColumns, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Delete("/api/v1/attendance/:id", attendanceController.Delete, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/attendance", attendanceController.GetStatistics, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/attendance/piechart", attendanceController.GetPieChartStatistics, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/attendance/barchart", attendanceController.GetBarChartStatistics, middleware.Authenticate(r.auth, auth.RoleAdmin))
	r.Get("/api/v1/attendance/graph", attendanceController.GetGraphStatistic, middleware.Authenticate(r.auth, auth.RoleAdmin))

	return r.Run(r.port)
}
