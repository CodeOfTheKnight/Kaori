package main

type endpoints string

const (
	endpointRoot  endpoints = "/"
	endpointGui   endpoints = "/KaoriGui/"
	endpointLogin endpoints = "/login"
	endpointAnime    endpoints = "/anime/{id:[0-9]+}"   //[GET] Parametri: IdAnilist
	endpointManga    endpoints = "/manga/{id:[0-9]+}"   //[GET] Parametri: IdAnilist
	endpointEpisode endpoints = "/episode"
	endpointChapter endpoints = "/chapter"

	//Endpoint per servizio
	endpointService endpoints = "/api/service/"
	serviceMusic    endpoints = "/music"   //[GET] Parametri: IdAnilist, [tipo](OP, ED, OST)

	//Endpoint utenti
	endpointUser endpoints = "/api/user/"
	userInfo     endpoints = "/info"     //[GET]
	userSettings endpoints = "/settings" //[GET]
	userBadge    endpoints = "/badge"    //[GET]

	//Endpoint auth
	endpointAuth      endpoints = "/api/auth/"
	authRefresh       endpoints = "/refresh"
	authLogin         endpoints = "/login"
	authSignUp        endpoints = "/signup"
	authConfirmSignUp endpoints = "/confirm"
	authRejectSignUp  endpoints = "/reject"
	authUserExist     endpoints = "/exist" //[GET] Parametri: email

	//Endpoint addData
	endpointAddData endpoints = "/api/addData/"
	addDataAnime    endpoints = "/anime" //[POST] Struttura dati da definire.
	addDataManga    endpoints = "/manga" //[POST] Struttura dati da definire.
	addDataMusic    endpoints = "/music" //[POST] Struttura dati da definire.

	//Endpoint lista anime/manga
	endpointUserList endpoints = "/api/list/userList" //[Get] Parametri: TokenAnilist, TokenServer, TipoLista
	endpointUserInfo endpoints = "/api/list/userinfo" //[Get] Parametri: TokenAnilist, TokenServer

	//Endpoint for tests
	endpointTest endpoints = "/api/test/"
	testFiles    endpoints = "/files/"

	//ENDPOINT ADMIN
	endpointAdmin  endpoints = "/api/admin/"
	adminConfig endpoints = "/config" //[GET]
	adminLogServer endpoints = "/log/server" //[GET] Parametri: ip, functions, user, level, msg, time
	adminLogConnection endpoints = "/log/connection"
	adminAnimeInsert endpoints = "/anime/insert" // [POST] Serve per caricare un anime nel database

		//ENDPOINT COMMAND
		adminCommand endpoints = "/command/"
		commandRestart endpoints = "/restart"
		commandShutdown endpoints = "/shutdown"
		commandForcedShutdown endpoints = "/forcedShutdown"

)

func (e endpoints) String() string {
	return string(e)
}
