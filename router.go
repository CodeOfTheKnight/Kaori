package main

import (
	"github.com/CodeOfTheKnight/Kaori/kaoriJwt"
	"github.com/CodeOfTheKnight/Kaori/kaoriMiddleware"
	"github.com/gorilla/mux"
	"net/http"
	"path"
)

//RouterInit create a new router with the routes and middleware already set up.
func RouterInit() *mux.Router {

	//Setting auth middleware
	userAuthMiddleware := NewAuthMiddlewarePerm(kaoriJwt.UserPerm)
	creatorAuthMiddleware := NewAuthMiddlewarePerm(kaoriJwt.CreatorPerm)
	testerAuthMiddleware := NewAuthMiddlewarePerm(kaoriJwt.TesterPerm)
	adminAuthMiddleware := NewAuthMiddlewarePerm(kaoriJwt.AdminPerm)

	//Creazione router
	router := mux.NewRouter()
	router.Use(kaoriMiddleware.EnableCors) //CORS middleware

	//Creazione subrouter per api di aggiunta dati
	routerAdd := router.PathPrefix(endpointAddData.String()).Subrouter()
	routerAdd.Use(userAuthMiddleware.authmiddleware)

	//Creazione subrouter per api di test
	routerTest := router.PathPrefix(endpointTest.String()).Subrouter()
	routerTest.Use(testerAuthMiddleware.authmiddleware)

	//Creazione subrouter per api di utente
	routerUser := router.PathPrefix(endpointUser.String()).Subrouter()
	routerUser.Use(userAuthMiddleware.authmiddleware)

	//Creazione subrouter per api di autenticazione
	routerAuth := router.PathPrefix(endpointAuth.String()).Subrouter()

	//Creazione subrouter per api di amministratore
	routerAdmin := router.PathPrefix(endpointAdmin.String()).Subrouter()
	routerAdmin.Use(adminAuthMiddleware.authmiddleware)

		//Creazione subrouter per api di amministratore relativo ai controlli
		routerCommand := routerAdmin.PathPrefix(adminCommand.String()).Subrouter()

	//Rotte base
	router.Handle(endpointRoot.String(), refreshMiddleware(http.HandlerFunc(serveIndex)))
	router.HandleFunc(endpointLogin.String(), serveLogin)
	router.PathPrefix("/KaoriGui/").Handler(http.StripPrefix("/KaoriGui/", http.FileServer(http.Dir(cfg.Server.Gui))))
	router.Handle(endpointAnime.String(), creatorAuthMiddleware.authmiddleware(http.HandlerFunc(ApiAnimeInsert))).Methods(http.MethodPut)
	router.Path(endpointAnime.String()).HandlerFunc(ApiServiceAnime).Methods(http.MethodGet)
	router.Handle(endpointManga.String(), creatorAuthMiddleware.authmiddleware(http.HandlerFunc(ApiMangaInsert))).Methods(http.MethodPut)
	router.Path(endpointManga.String()).HandlerFunc(ApiServiceManga).Methods(http.MethodGet)

	//Rotte API AddData
	routerAdd.Path(addDataMusic.String()).HandlerFunc(ApiAddMusic).Methods(http.MethodPost)

	//Rotte API Auth
	routerAuth.Path(authLogin.String()).HandlerFunc(ApiLogin).Methods(http.MethodPost)
	routerAuth.Path(authRefresh.String()).HandlerFunc(ApiRefresh).Methods(http.MethodGet)
	routerAuth.Path(authSignUp.String()).HandlerFunc(ApiSignUp).Methods(http.MethodPost)
	routerAuth.Path(authConfirmSignUp.String()).HandlerFunc(ApiConfirmSignUp).Methods(http.MethodGet)
	routerAuth.Path(authUserExist.String()).HandlerFunc(ApiUserExist).Methods(http.MethodGet)

	//Rotte API test
	routerTest.PathPrefix(testFiles.String()).Handler(
		http.StripPrefix(
			path.Join(
				endpointTest.String(),
				testFiles.String(),
			),
			http.FileServer(http.Dir(cfg.Server.Test)),
		),
	).Methods(http.MethodGet, http.MethodOptions)

	//Rotte API user
	routerUser.Path(userInfo.String()).HandlerFunc(ApiUserInfo).Methods(http.MethodGet)
	routerUser.Path(userSettings.String()).HandlerFunc(ApiSettingsGet).Methods(http.MethodGet)
	routerUser.Path(userSettings.String()).HandlerFunc(ApiSettingsSet).Methods(http.MethodPut)

	//Rotte API admin
	routerAdmin.Path(adminConfig.String()).HandlerFunc(ApiConfigGet).Methods(http.MethodGet)
	routerAdmin.Path(adminConfig.String()).HandlerFunc(ApiConfigSet).Methods(http.MethodPut)
	routerAdmin.Path(adminLogServer.String()).HandlerFunc(ApiLogServer).Methods(http.MethodGet, http.MethodPost)
	routerAdmin.Path(adminLogConnection.String()).HandlerFunc(ApiLogConnection).Methods(http.MethodGet, http.MethodPost)

		//Rotte API command
		routerCommand.Path(commandRestart.String()).HandlerFunc(ApiCommandRestart).Methods(http.MethodGet)
		routerCommand.Path(commandShutdown.String()).HandlerFunc(ApiCommandShutdown).Methods(http.MethodGet)
		routerCommand.Path(commandForcedShutdown.String()).HandlerFunc(ApiCommandForcedShutdown).Methods(http.MethodGet)

	return router
}
