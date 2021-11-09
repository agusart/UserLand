package main

import (
	"github.com/go-chi/chi/v5"
	"userland/api/auth"
	"userland/api/me"
	"userland/api/middleware"
	"userland/api/session"
	"userland/store/broker"
	"userland/store/postgres"
	"userland/store/redis"
)

var router *chi.Mux

func InitRouter(
	jwtHandler middleware.JwtHandlerInterface,
	cache redis.CacheInterface,
	fileHelper me.FileHelperInterface,
	userStore postgres.UserStoreInterface,
	authStore redis.AuthStoreInterface,
	sessionStore postgres.SessionStoreInterface,
	tfaStore postgres.TfaStoreInterface,
	broker broker.MessageBrokerInterface,
) *chi.Mux {
	if router != nil {
		return router
	}
	router = chi.NewMux()
	authMiddleware := middleware.NewAuthMiddleware(jwtHandler, cache)

	router.Get("/asset/{filename}", me.ShowImages(fileHelper))

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.Register(userStore, authStore))
		r.Post("/verification", auth.RequestVerification(userStore, authStore))
		r.Get("/verify/{token}", auth.VerifyRegister(userStore, authStore))

		r.With(middleware.RequestMustBeJsonMiddleware).Post("/password/forgot", auth.ForgetPassword(userStore, authStore))
		r.With(middleware.RequestMustBeJsonMiddleware).Post("/password/reset", auth.ResetPassword(userStore, authStore))
		r.With(middleware.RequestMustBeJsonMiddleware).Post("/login", auth.Login(userStore, jwtHandler, sessionStore, authStore, broker))
		r.With(middleware.RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware).Post("/tfa/verify", auth.VerifyTfa(jwtHandler, userStore, tfaStore))
		r.With(middleware.RequestMustBeJsonMiddleware, authMiddleware.UserAuthMiddleware).Post("/tfa/bypass", auth.BypassTfa(jwtHandler, tfaStore))
	})

	router.Route("/me", func(r chi.Router) {
		r.Use(authMiddleware.UserAuthMiddleware, middleware.TfaRequiredMiddleware)

		r.Get("/session", session.ListSession(sessionStore))
		r.Delete("/session", session.EndSession(sessionStore, cache))
		r.Delete("/session/other", session.EndAllOtherSessions(sessionStore, cache))
		r.Get("/session/refresh-token", session.RefreshToken(jwtHandler))
		r.Get("/session/access-token", session.NewAccessToken(jwtHandler))

		r.Get("/", me.UserDetail(userStore))
		r.With(middleware.RequestMustBeJsonMiddleware).Post("/", me.UpdateUserDetail(userStore))

		r.Get("/email", me.GetCurrentEmailAddress(userStore))
		r.With(middleware.RequestMustBeJsonMiddleware).Post("/email", me.UpdateUserEmailRequest(cache))
		r.Get("/email/verify/{verifyToken}", me.UpdateUserEmail(userStore, cache))
		r.With(middleware.RequestMustBeJsonMiddleware).Post("/password", me.UpdateUserPassword(userStore))

		r.Get("/tfa/status", me.GetCurrentTfaStatus(userStore))
		r.Get("/tfa/enroll", me.SetupTfa(userStore))
		r.With(middleware.RequestMustBeJsonMiddleware).Post("/tfa/enroll", me.ActivateTfa(tfaStore))
		r.Post("/tfa/remove", me.RemoveTfa(userStore, tfaStore))

		r.With(middleware.RequestMustBeJsonMiddleware).Post("/delete", me.DeleteAccount(userStore))


		r.Delete("/picture", me.DeleteImages(userStore))
		r.Post("/picture", me.UploadPhoto(userStore, fileHelper))

	})

	return router
}
